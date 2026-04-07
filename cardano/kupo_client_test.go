package cardano

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer creates a Kupo mock that serves canned UTxO responses.
func newTestServer(t *testing.T, utxos []kupoUTxO) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/health":
			json.NewEncoder(w).Encode(kupoHealth{
				ConnectionStatus: "connected",
				Version:          "test",
			})
		default:
			json.NewEncoder(w).Encode(utxos)
		}
	}))
}

func TestKupoClient_Health(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(kupoHealth{ConnectionStatus: "connected"})
	}))
	defer srv.Close()

	client, err := NewKupoClient(KupoConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Health(context.Background()); err != nil {
		t.Fatalf("Health() = %v", err)
	}
}

func TestKupoClient_Health_Disconnected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(kupoHealth{ConnectionStatus: "disconnected"})
	}))
	defer srv.Close()

	client, err := NewKupoClient(KupoConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Health(context.Background()); err == nil {
		t.Fatal("Health() should fail when disconnected")
	}
}

var testUTxOs = []kupoUTxO{
	{
		TransactionID: "tx1",
		OutputIndex:   0,
		Address:       "addr_test1",
		Value: kupoValue{
			Coins: json.Number("5000000"),
			Assets: map[string]uint64{
				"policyA.tokenX": 100,
			},
		},
	},
	{
		TransactionID: "tx2",
		OutputIndex:   1,
		Address:       "addr_test1",
		Value: kupoValue{
			Coins: json.Number("3000000"),
			Assets: map[string]uint64{
				"policyA.tokenX": 50,
				"policyB.tokenY": 200,
			},
		},
	},
}

func TestKupoClient_UTxOs(t *testing.T) {
	srv := newTestServer(t, testUTxOs)
	defer srv.Close()

	client, err := NewKupoClient(KupoConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	utxos, err := client.UTxOs(context.Background(), "addr_test1")
	if err != nil {
		t.Fatalf("UTxOs() = %v", err)
	}
	if len(utxos) != 2 {
		t.Fatalf("got %d UTxOs, want 2", len(utxos))
	}
	assertBigInt(t, "utxo[0].Value.Coin", utxos[0].Value.Coin, bi(5_000_000))
}

func TestKupoClient_GetAllBalances(t *testing.T) {
	srv := newTestServer(t, testUTxOs)
	defer srv.Close()

	client, err := NewKupoClient(KupoConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	balances, err := client.GetAllBalances(context.Background(), "addr_test1")
	if err != nil {
		t.Fatalf("GetAllBalances() = %v", err)
	}

	// Expect: lovelace=8_000_000, policyA.tokenX=150, policyB.tokenY=200
	balMap := make(map[string]Balance)
	for _, b := range balances {
		balMap[b.Asset.String()] = b
	}

	assertBigInt(t, "lovelace", balMap["lovelace"].Quantity, bi(8_000_000))
	if b := balMap["lovelace"]; b.UTxOCount != 2 {
		t.Errorf("lovelace UTxOCount = %d, want 2", b.UTxOCount)
	}
	assertBigInt(t, "policyA.tokenX", balMap["policyA.tokenX"].Quantity, bi(150))
	assertBigInt(t, "policyB.tokenY", balMap["policyB.tokenY"].Quantity, bi(200))
}

func TestKupoClient_GetBalanceOfCoin_Lovelace(t *testing.T) {
	srv := newTestServer(t, testUTxOs)
	defer srv.Close()

	client, err := NewKupoClient(KupoConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	bal, err := client.GetBalanceOfCoin(context.Background(), "addr_test1", Asset{})
	if err != nil {
		t.Fatalf("GetBalanceOfCoin(lovelace) = %v", err)
	}
	assertBigInt(t, "lovelace", bal.Quantity, bi(8_000_000))
	if bal.UTxOCount != 2 {
		t.Errorf("UTxOCount = %d, want 2", bal.UTxOCount)
	}
}

func TestKupoClient_GetBalanceOfCoin_NativeToken(t *testing.T) {
	srv := newTestServer(t, testUTxOs)
	defer srv.Close()

	client, err := NewKupoClient(KupoConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	asset := Asset{PolicyID: "policyA", TokenName: "tokenX"}
	bal, err := client.GetBalanceOfCoin(context.Background(), "addr_test1", asset)
	if err != nil {
		t.Fatalf("GetBalanceOfCoin(policyA.tokenX) = %v", err)
	}
	assertBigInt(t, "policyA.tokenX", bal.Quantity, bi(150))
}

func TestKupoClient_EmptyAddress(t *testing.T) {
	srv := newTestServer(t, nil)
	defer srv.Close()

	client, err := NewKupoClient(KupoConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	balances, err := client.GetAllBalances(context.Background(), "addr_empty")
	if err != nil {
		t.Fatalf("GetAllBalances() = %v", err)
	}
	if len(balances) != 0 {
		t.Errorf("got %d balances, want 0", len(balances))
	}
}

func TestNewKupoClient_MissingURL(t *testing.T) {
	_, err := NewKupoClient(KupoConfig{})
	if err == nil {
		t.Fatal("NewKupoClient should fail with empty BaseURL")
	}
}

func TestKupoClient_InterfaceCompliance(t *testing.T) {
	var _ Client = (*KupoClient)(nil)
}

func TestKupoClient_UTxOsByTxIns(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.Contains(path, "0@tx1"):
			json.NewEncoder(w).Encode([]kupoUTxO{testUTxOs[0]})
		case strings.Contains(path, "1@tx2"):
			json.NewEncoder(w).Encode([]kupoUTxO{testUTxOs[1]})
		default:
			json.NewEncoder(w).Encode([]kupoUTxO{})
		}
	}))
	defer srv.Close()

	client, err := NewKupoClient(KupoConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	txIns := []TxIn{
		{TxHash: "tx1", Index: 0},
		{TxHash: "tx2", Index: 1},
	}
	utxos, err := client.UTxOsByTxIns(context.Background(), txIns)
	if err != nil {
		t.Fatalf("UTxOsByTxIns() = %v", err)
	}
	if len(utxos) != 2 {
		t.Fatalf("got %d UTxOs, want 2", len(utxos))
	}
	if utxos[0].Input.TxHash != "tx1" {
		t.Errorf("utxo[0].TxHash = %s, want tx1", utxos[0].Input.TxHash)
	}
	if utxos[1].Input.TxHash != "tx2" {
		t.Errorf("utxo[1].TxHash = %s, want tx2", utxos[1].Input.TxHash)
	}
}

func TestKupoClient_UTxOsByTxIns_Missing(t *testing.T) {
	srv := newTestServer(t, nil)
	defer srv.Close()

	client, err := NewKupoClient(KupoConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	txIns := []TxIn{{TxHash: "nonexistent", Index: 0}}
	utxos, err := client.UTxOsByTxIns(context.Background(), txIns)
	if err != nil {
		t.Fatalf("UTxOsByTxIns() = %v", err)
	}
	if len(utxos) != 0 {
		t.Errorf("got %d UTxOs, want 0 for missing TxIn", len(utxos))
	}
}

func TestKupoClient_UTxOsByPaymentCredential(t *testing.T) {
	cred := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef01"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/matches/"+cred) {
			json.NewEncoder(w).Encode(testUTxOs)
		} else {
			json.NewEncoder(w).Encode([]kupoUTxO{})
		}
	}))
	defer srv.Close()

	client, err := NewKupoClient(KupoConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	utxos, err := client.UTxOsByPaymentCredential(context.Background(), cred)
	if err != nil {
		t.Fatalf("UTxOsByPaymentCredential() = %v", err)
	}
	if len(utxos) != 2 {
		t.Fatalf("got %d UTxOs, want 2", len(utxos))
	}
	assertBigInt(t, "utxo[0].Coin", utxos[0].Value.Coin, bi(5_000_000))
}

func TestKupoClient_UTxOsByPaymentCredential_Empty(t *testing.T) {
	srv := newTestServer(t, nil)
	defer srv.Close()

	client, err := NewKupoClient(KupoConfig{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	utxos, err := client.UTxOsByPaymentCredential(context.Background(), "deadbeef")
	if err != nil {
		t.Fatalf("UTxOsByPaymentCredential() = %v", err)
	}
	if len(utxos) != 0 {
		t.Errorf("got %d UTxOs, want 0", len(utxos))
	}
}
