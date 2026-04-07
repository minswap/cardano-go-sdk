package cardano

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

// Asset identifies a native token on Cardano by its policy ID and token name.
// The zero value represents ADA (lovelace).
type Asset struct {
	PolicyID  string `json:"policyId"`
	TokenName string `json:"tokenName"`
}

// IsLovelace reports whether this asset represents ADA (lovelace).
func (a Asset) IsLovelace() bool {
	return a.PolicyID == "" && a.TokenName == ""
}

// String returns "lovelace" for ADA, or "policyId.tokenName" for native tokens.
func (a Asset) String() string {
	if a.IsLovelace() {
		return "lovelace"
	}
	if a.TokenName == "" {
		return a.PolicyID
	}
	return a.PolicyID + "." + a.TokenName
}

// ParseAsset parses a "policyId.tokenName" string into an Asset.
// Returns the lovelace Asset for "lovelace" or empty string.
func ParseAsset(s string) Asset {
	if s == "" || s == "lovelace" {
		return Asset{}
	}
	parts := strings.SplitN(s, ".", 2)
	if len(parts) == 1 {
		return Asset{PolicyID: parts[0]}
	}
	return Asset{PolicyID: parts[0], TokenName: parts[1]}
}

// Value represents a Cardano multi-asset value: lovelace plus optional native tokens.
type Value struct {
	Coin       *big.Int                       `json:"coin"`
	MultiAsset map[string]map[string]*big.Int `json:"multiAsset,omitempty"`
}

// NewValue creates a Value with the given lovelace amount and no multi-asset.
func NewValue(coin int64) Value {
	return Value{Coin: big.NewInt(coin)}
}

// Get returns the quantity of a specific asset in this value.
// Returns a new zero big.Int if the asset is not present.
func (v Value) Get(asset Asset) *big.Int {
	if asset.IsLovelace() {
		if v.Coin == nil {
			return new(big.Int)
		}
		return new(big.Int).Set(v.Coin)
	}
	if v.MultiAsset == nil {
		return new(big.Int)
	}
	tokens, ok := v.MultiAsset[asset.PolicyID]
	if !ok {
		return new(big.Int)
	}
	qty := tokens[asset.TokenName]
	if qty == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(qty)
}

// Assets returns all non-zero asset entries as (Asset, quantity) pairs.
// Lovelace is included as the first entry if non-zero.
func (v Value) Assets() []AssetQuantity {
	var result []AssetQuantity
	if v.Coin != nil && v.Coin.Sign() > 0 {
		result = append(result, AssetQuantity{Asset: Asset{}, Quantity: new(big.Int).Set(v.Coin)})
	}
	for policyID, tokens := range v.MultiAsset {
		for tokenName, qty := range tokens {
			if qty != nil && qty.Sign() > 0 {
				result = append(result, AssetQuantity{
					Asset:    Asset{PolicyID: policyID, TokenName: tokenName},
					Quantity: new(big.Int).Set(qty),
				})
			}
		}
	}
	return result
}

// Add returns a new Value that is the sum of v and other.
func (v Value) Add(other Value) Value {
	coin := new(big.Int)
	if v.Coin != nil {
		coin.Add(coin, v.Coin)
	}
	if other.Coin != nil {
		coin.Add(coin, other.Coin)
	}

	result := Value{
		Coin:       coin,
		MultiAsset: make(map[string]map[string]*big.Int),
	}
	// Copy v's multi-asset
	for pid, tokens := range v.MultiAsset {
		result.MultiAsset[pid] = make(map[string]*big.Int)
		for tn, qty := range tokens {
			result.MultiAsset[pid][tn] = new(big.Int).Set(qty)
		}
	}
	// Add other's multi-asset
	for pid, tokens := range other.MultiAsset {
		if result.MultiAsset[pid] == nil {
			result.MultiAsset[pid] = make(map[string]*big.Int)
		}
		for tn, qty := range tokens {
			existing := result.MultiAsset[pid][tn]
			if existing == nil {
				existing = new(big.Int)
			}
			result.MultiAsset[pid][tn] = existing.Add(existing, qty)
		}
	}
	return result
}

// AssetQuantity pairs an Asset with its quantity.
type AssetQuantity struct {
	Asset    Asset    `json:"asset"`
	Quantity *big.Int `json:"quantity"`
}

// Balance represents the total balance for a specific asset type,
// analogous to Sui's Balance struct.
type Balance struct {
	Asset    Asset    `json:"asset"`
	Quantity *big.Int `json:"quantity"`
	// UTxOCount is the number of UTxOs contributing to this balance.
	UTxOCount int `json:"utxoCount"`
}

// TxIn identifies a specific transaction output (UTxO reference).
type TxIn struct {
	TxHash string `json:"txHash"`
	Index  int    `json:"index"`
}

func (t TxIn) String() string {
	return fmt.Sprintf("%d@%s", t.Index, t.TxHash)
}

// UTxO represents an unspent transaction output on Cardano.
type UTxO struct {
	Input   TxIn   `json:"input"`
	Address string `json:"address"`
	Value   Value  `json:"value"`
}

// kupoUTxO is the raw JSON shape returned by Kupo's /matches endpoint.
type kupoUTxO struct {
	TransactionID string    `json:"transaction_id"`
	OutputIndex   int       `json:"output_index"`
	Address       string    `json:"address"`
	Value         kupoValue `json:"value"`
}

type kupoValue struct {
	Coins  json.Number       `json:"coins"`
	Assets map[string]uint64 `json:"assets,omitempty"`
}

// toUTxO converts a Kupo response into our domain type.
func (k kupoUTxO) toUTxO() UTxO {
	coin := new(big.Int)
	coin.UnmarshalJSON([]byte(k.Value.Coins.String()))

	v := Value{
		Coin: coin,
	}
	if len(k.Value.Assets) > 0 {
		v.MultiAsset = make(map[string]map[string]*big.Int)
		for key, qty := range k.Value.Assets {
			asset := ParseAsset(key)
			if v.MultiAsset[asset.PolicyID] == nil {
				v.MultiAsset[asset.PolicyID] = make(map[string]*big.Int)
			}
			v.MultiAsset[asset.PolicyID][asset.TokenName] = new(big.Int).SetUint64(qty)
		}
	}
	return UTxO{
		Input: TxIn{
			TxHash: k.TransactionID,
			Index:  k.OutputIndex,
		},
		Address: k.Address,
		Value:   v,
	}
}

// kupoHealth is the JSON shape returned by Kupo's /health endpoint.
type kupoHealth struct {
	ConnectionStatus     string      `json:"connection_status"`
	MostRecentNodeTip    json.Number `json:"most_recent_node_tip"`
	MostRecentCheckpoint json.Number `json:"most_recent_checkpoint"`
	Version              string      `json:"version"`
}
