package cardano

import (
	"encoding/json"
	"math/big"
	"testing"
)

func bi(n int64) *big.Int { return big.NewInt(n) }

func assertBigInt(t *testing.T, label string, got, want *big.Int) {
	t.Helper()
	if got.Cmp(want) != 0 {
		t.Errorf("%s = %s, want %s", label, got, want)
	}
}

func TestAsset_String(t *testing.T) {
	tests := []struct {
		asset Asset
		want  string
	}{
		{Asset{}, "lovelace"},
		{Asset{PolicyID: "abc123"}, "abc123"},
		{Asset{PolicyID: "abc123", TokenName: "TOKEN"}, "abc123.TOKEN"},
	}
	for _, tt := range tests {
		if got := tt.asset.String(); got != tt.want {
			t.Errorf("Asset%+v.String() = %q, want %q", tt.asset, got, tt.want)
		}
	}
}

func TestAsset_IsLovelace(t *testing.T) {
	lovelace := Asset{}
	if !lovelace.IsLovelace() {
		t.Error("zero Asset should be lovelace")
	}
	token := Asset{PolicyID: "abc"}
	if token.IsLovelace() {
		t.Error("asset with policyID should not be lovelace")
	}
}

func TestParseAsset(t *testing.T) {
	tests := []struct {
		input string
		want  Asset
	}{
		{"lovelace", Asset{}},
		{"", Asset{}},
		{"abc123", Asset{PolicyID: "abc123"}},
		{"abc123.TOKEN", Asset{PolicyID: "abc123", TokenName: "TOKEN"}},
		{"abc123.TO.KEN", Asset{PolicyID: "abc123", TokenName: "TO.KEN"}},
	}
	for _, tt := range tests {
		got := ParseAsset(tt.input)
		if got != tt.want {
			t.Errorf("ParseAsset(%q) = %+v, want %+v", tt.input, got, tt.want)
		}
	}
}

func TestValue_Get(t *testing.T) {
	v := Value{
		Coin: bi(5_000_000),
		MultiAsset: map[string]map[string]*big.Int{
			"policy1": {"tokenA": bi(100), "tokenB": bi(200)},
		},
	}

	assertBigInt(t, "Get(lovelace)", v.Get(Asset{}), bi(5_000_000))
	assertBigInt(t, "Get(policy1.tokenA)", v.Get(Asset{PolicyID: "policy1", TokenName: "tokenA"}), bi(100))
	assertBigInt(t, "Get(missing)", v.Get(Asset{PolicyID: "missing"}), bi(0))
}

func TestValue_Add(t *testing.T) {
	a := Value{
		Coin: bi(1_000_000),
		MultiAsset: map[string]map[string]*big.Int{
			"p1": {"t1": bi(10)},
		},
	}
	b := Value{
		Coin: bi(2_000_000),
		MultiAsset: map[string]map[string]*big.Int{
			"p1": {"t1": bi(5), "t2": bi(20)},
			"p2": {"t3": bi(30)},
		},
	}

	result := a.Add(b)
	assertBigInt(t, "Coin", result.Coin, bi(3_000_000))
	assertBigInt(t, "p1.t1", result.MultiAsset["p1"]["t1"], bi(15))
	assertBigInt(t, "p1.t2", result.MultiAsset["p1"]["t2"], bi(20))
	assertBigInt(t, "p2.t3", result.MultiAsset["p2"]["t3"], bi(30))
}

func TestValue_Assets(t *testing.T) {
	v := Value{
		Coin: bi(1_000_000),
		MultiAsset: map[string]map[string]*big.Int{
			"p1": {"t1": bi(100)},
		},
	}

	assets := v.Assets()
	if len(assets) != 2 {
		t.Fatalf("Assets() returned %d entries, want 2", len(assets))
	}

	hasLovelace := false
	hasToken := false
	for _, aq := range assets {
		if aq.Asset.IsLovelace() && aq.Quantity.Cmp(bi(1_000_000)) == 0 {
			hasLovelace = true
		}
		if aq.Asset.PolicyID == "p1" && aq.Asset.TokenName == "t1" && aq.Quantity.Cmp(bi(100)) == 0 {
			hasToken = true
		}
	}
	if !hasLovelace {
		t.Error("missing lovelace entry")
	}
	if !hasToken {
		t.Error("missing p1.t1 entry")
	}
}

func TestKupoUTxO_toUTxO(t *testing.T) {
	raw := kupoUTxO{
		TransactionID: "abc123",
		OutputIndex:   0,
		Address:       "addr_test1qz...",
		Value: kupoValue{
			Coins: json.Number("2000000"),
			Assets: map[string]uint64{
				"policyA.tokenX": 500,
				"policyA.tokenY": 300,
				"policyB.tokenZ": 100,
			},
		},
	}

	utxo := raw.toUTxO()
	if utxo.Input.TxHash != "abc123" || utxo.Input.Index != 0 {
		t.Errorf("unexpected input: %+v", utxo.Input)
	}
	assertBigInt(t, "Coin", utxo.Value.Coin, bi(2_000_000))
	assertBigInt(t, "policyA.tokenX", utxo.Value.MultiAsset["policyA"]["tokenX"], bi(500))
	assertBigInt(t, "policyB.tokenZ", utxo.Value.MultiAsset["policyB"]["tokenZ"], bi(100))
}
