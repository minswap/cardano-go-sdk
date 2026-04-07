package cardano

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"
)

// KupoConfig holds configuration for connecting to a Kupo instance.
type KupoConfig struct {
	// BaseURL is the Kupo HTTP endpoint (e.g. "http://localhost:1442").
	BaseURL string

	// HTTPClient is an optional custom HTTP client. If nil, a default
	// client with a 30-second timeout is used.
	HTTPClient *http.Client
}

// KupoClient implements Client by querying a Kupo chain indexer.
type KupoClient struct {
	baseURL string
	http    *http.Client
}

// NewKupoClient creates a new Kupo-backed Client.
func NewKupoClient(cfg KupoConfig) (*KupoClient, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("cardano: kupo base URL is required")
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &KupoClient{
		baseURL: baseURL,
		http:    httpClient,
	}, nil
}

// Health checks whether Kupo is connected and synced.
func (c *KupoClient) Health(ctx context.Context) error {
	var health kupoHealth
	if err := c.get(ctx, "/health", &health); err != nil {
		return fmt.Errorf("kupo health check: %w", err)
	}
	if health.ConnectionStatus != "connected" {
		return fmt.Errorf("kupo not connected: status=%s", health.ConnectionStatus)
	}
	return nil
}

// UTxOs returns all unspent transaction outputs at the given address.
func (c *KupoClient) UTxOs(ctx context.Context, address string) ([]UTxO, error) {
	var raw []kupoUTxO
	path := fmt.Sprintf("/matches/%s?unspent", address)
	if err := c.get(ctx, path, &raw); err != nil {
		return nil, fmt.Errorf("kupo utxos at %s: %w", address, err)
	}

	utxos := make([]UTxO, len(raw))
	for i, r := range raw {
		utxos[i] = r.toUTxO()
	}
	return utxos, nil
}

// UTxOsByTxIns fetches specific unspent outputs by their transaction input references.
// Each TxIn is queried individually via Kupo's /matches/{index}@{txHash}?unspent pattern.
// Missing or spent inputs are silently omitted from the result.
func (c *KupoClient) UTxOsByTxIns(ctx context.Context, txIns []TxIn) ([]UTxO, error) {
	var result []UTxO
	for _, txIn := range txIns {
		var raw []kupoUTxO
		path := fmt.Sprintf("/matches/%d@%s?unspent", txIn.Index, txIn.TxHash)
		if err := c.get(ctx, path, &raw); err != nil {
			return nil, fmt.Errorf("kupo utxo %s: %w", txIn, err)
		}
		for _, r := range raw {
			result = append(result, r.toUTxO())
		}
	}
	return result, nil
}

// UTxOsByPaymentCredential returns all unspent outputs locked by the given
// payment credential hash (hex-encoded).
// Uses Kupo's /matches/{credential}/*?unspent pattern.
func (c *KupoClient) UTxOsByPaymentCredential(ctx context.Context, credential string) ([]UTxO, error) {
	var raw []kupoUTxO
	path := fmt.Sprintf("/matches/%s/*?unspent", credential)
	if err := c.get(ctx, path, &raw); err != nil {
		return nil, fmt.Errorf("kupo utxos by credential %s: %w", credential, err)
	}

	utxos := make([]UTxO, len(raw))
	for i, r := range raw {
		utxos[i] = r.toUTxO()
	}
	return utxos, nil
}

// GetAllBalances returns one Balance entry per distinct asset held at the address.
func (c *KupoClient) GetAllBalances(ctx context.Context, address string) ([]Balance, error) {
	utxos, err := c.UTxOs(ctx, address)
	if err != nil {
		return nil, err
	}
	return aggregateBalances(utxos), nil
}

// GetBalanceOfCoin returns the balance of a specific asset at the address.
// Pass Asset{} (zero value) for ADA/lovelace.
func (c *KupoClient) GetBalanceOfCoin(ctx context.Context, address string, asset Asset) (*Balance, error) {
	// For native tokens, Kupo supports filtering: /matches/{addr}/{policyId.tokenName}
	if !asset.IsLovelace() {
		return c.getFilteredBalance(ctx, address, asset)
	}

	// For lovelace, we must fetch all UTxOs and sum.
	utxos, err := c.UTxOs(ctx, address)
	if err != nil {
		return nil, err
	}
	bal := Balance{Asset: asset, Quantity: new(big.Int)}
	for _, u := range utxos {
		bal.Quantity.Add(bal.Quantity, u.Value.Coin)
		bal.UTxOCount++
	}
	return &bal, nil
}

// getFilteredBalance uses Kupo's asset filter to query UTxOs holding a specific token.
func (c *KupoClient) getFilteredBalance(ctx context.Context, address string, asset Asset) (*Balance, error) {
	var raw []kupoUTxO
	path := fmt.Sprintf("/matches/%s?unspent&policy_id=%s", address, asset.PolicyID)
	if asset.TokenName != "" {
		path += fmt.Sprintf("&asset_name=%s", asset.TokenName)
	}
	if err := c.get(ctx, path, &raw); err != nil {
		return nil, fmt.Errorf("kupo balance of %s at %s: %w", asset, address, err)
	}

	bal := Balance{Asset: asset, Quantity: new(big.Int)}
	for _, r := range raw {
		utxo := r.toUTxO()
		bal.Quantity.Add(bal.Quantity, utxo.Value.Get(asset))
		bal.UTxOCount++
	}
	return &bal, nil
}

// get performs an HTTP GET and decodes the JSON response into dst.
func (c *KupoClient) get(ctx context.Context, path string, dst any) error {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

// aggregateBalances sums UTxO values into per-asset Balance entries.
func aggregateBalances(utxos []UTxO) []Balance {
	type key struct {
		policyID  string
		tokenName string
	}
	counts := make(map[key]*big.Int)
	utxoCounts := make(map[key]int)

	addTo := func(k key, qty *big.Int) {
		if counts[k] == nil {
			counts[k] = new(big.Int)
		}
		counts[k].Add(counts[k], qty)
		utxoCounts[k]++
	}

	for _, u := range utxos {
		addTo(key{}, u.Value.Coin)

		for pid, tokens := range u.Value.MultiAsset {
			for tn, qty := range tokens {
				addTo(key{policyID: pid, tokenName: tn}, qty)
			}
		}
	}

	balances := make([]Balance, 0, len(counts))
	for k, qty := range counts {
		balances = append(balances, Balance{
			Asset:     Asset{PolicyID: k.policyID, TokenName: k.tokenName},
			Quantity:  qty,
			UTxOCount: utxoCounts[k],
		})
	}
	return balances
}
