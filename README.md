# Cardano Go SDK

A Go SDK for querying the Cardano blockchain.

## Installation

```bash
go get github.com/minswap/cardano-go-sdk
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/minswap/cardano-go-sdk/cardano"
)

func main() {
	client, err := cardano.NewKupoClient(cardano.KupoConfig{
		BaseURL: "http://localhost:1442",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	addr := "addr1q..."

	// Get all balances (ADA + native tokens)
	balances, err := client.GetAllBalances(ctx, addr)
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range balances {
		fmt.Printf("%s: %s (across %d UTxOs)\n", b.Asset, b.Quantity, b.UTxOCount)
	}

	// Get balance of a specific token
	bal, err := client.GetBalanceOfCoin(ctx, addr, cardano.Asset{
		PolicyID:  "abc123...",
		TokenName: "MIN",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("MIN balance: %s\n", bal.Quantity)

	// Get all UTxOs at an address
	utxos, err := client.UTxOs(ctx, addr)
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range utxos {
		fmt.Printf("UTxO %s: %s lovelace\n", u.Input, u.Value.Coin)
	}

	// Get UTxOs by specific transaction inputs
	utxos, err = client.UTxOsByTxIns(ctx, []cardano.TxIn{
		{TxHash: "abcdef...", Index: 0},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Get UTxOs by payment credential
	utxos, err = client.UTxOsByPaymentCredential(ctx, "abcdef0123456789...")
	if err != nil {
		log.Fatal(err)
	}
}
```

## Backend: Kupo

This SDK uses [Kupo](https://cardanosolutions.github.io/kupo/) as its chain indexer backend. Kupo must be running and synced to query chain state.

### Client Interface

```go
type Client interface {
	GetAllBalances(ctx context.Context, address string) ([]Balance, error)
	GetBalanceOfCoin(ctx context.Context, address string, asset Asset) (*Balance, error)
	UTxOs(ctx context.Context, address string) ([]UTxO, error)
	UTxOsByTxIns(ctx context.Context, txIns []TxIn) ([]UTxO, error)
	UTxOsByPaymentCredential(ctx context.Context, credential string) ([]UTxO, error)
	Health(ctx context.Context) error
}
```

### Types

All value amounts use `*big.Int` to handle Cardano's arbitrary-precision token quantities.

```go
// Value represents lovelace + multi-asset tokens
type Value struct {
	Coin       *big.Int
	MultiAsset map[string]map[string]*big.Int
}

// Balance for a specific asset across UTxOs
type Balance struct {
	Asset     Asset
	Quantity  *big.Int
	UTxOCount int
}
```

## Development

```bash
go test ./...
go vet ./...
```
