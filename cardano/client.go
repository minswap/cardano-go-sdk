// Package cardano provides a Go SDK for querying the Cardano blockchain.
package cardano

import "context"

// Client defines the interface for querying Cardano chain state.
// Implementations may use different backends (Kupo, Blockfrost, etc.).
type Client interface {
	// GetAllBalances returns the total balance for every asset held at the
	// given address, including lovelace and all native tokens.
	GetAllBalances(ctx context.Context, address string) ([]Balance, error)

	// GetBalanceOfCoin returns the total balance of a specific asset at the
	// given address. For ADA/lovelace, pass an empty Asset{}.
	GetBalanceOfCoin(ctx context.Context, address string, asset Asset) (*Balance, error)

	// UTxOs returns all unspent transaction outputs at the given address.
	UTxOs(ctx context.Context, address string) ([]UTxO, error)

	// UTxOsByTxIns returns the unspent outputs identified by the given
	// transaction inputs. Missing or already-spent inputs are silently omitted.
	UTxOsByTxIns(ctx context.Context, txIns []TxIn) ([]UTxO, error)

	// UTxOsByPaymentCredential returns all unspent outputs locked by the
	// given payment credential hash (hex-encoded, 28 bytes / 56 chars).
	UTxOsByPaymentCredential(ctx context.Context, credential string) ([]UTxO, error)

	// Health checks whether the backend is connected and synced.
	Health(ctx context.Context) error
}
