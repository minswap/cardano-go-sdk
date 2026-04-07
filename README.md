# Cardano Go SDK

A Go SDK for querying the Cardano blockchain, inspired by [sui-go-sdk](https://github.com/block-vision/sui-go-sdk).

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
		fmt.Printf("%s: %d (across %d UTxOs)\n", b.Asset, b.Quantity, b.UTxOCount)
	}

	// Get balance of a specific token
	bal, err := client.GetBalanceOfCoin(ctx, addr, cardano.Asset{
		PolicyID:  "abc123...",
		TokenName: "MIN",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("MIN balance: %d\n", bal.Quantity)

	// Get all UTxOs at an address
	utxos, err := client.UTxOs(ctx, addr)
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range utxos {
		fmt.Printf("UTxO %s: %d lovelace\n", u.Input, u.Value.Coin)
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
	Health(ctx context.Context) error
}
```

## Development

```bash
go test ./...
go vet ./...
```
