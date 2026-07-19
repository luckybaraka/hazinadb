// Package port declares the boundaries of the hexagon: the interfaces that
// connect the core to the outside world.
//
//   - Inbound (driving) ports  -> what the core OFFERS to callers.
//   - Outbound (driven) ports  -> what the core NEEDS from infrastructure.
//
// Adapters implement these interfaces; the core only ever refers to them.
package port

import (
	"context"

	"github.com/luckybaraka/hazinadb/internal/core/domain"
)

// KeyValueStore is the inbound (driving) port.
//
// It is the public API of HazinaDB's core. Driving adapters — the HTTP server
// today, a gRPC server or CLI tomorrow — depend on this interface and never on
// a concrete service type.
type KeyValueStore interface {
	// Get returns the value stored under key, or domain.ErrKeyNotFound.
	Get(ctx context.Context, key domain.Key) (domain.Value, error)
	// Put stores value under key, overwriting any existing value.
	Put(ctx context.Context, key domain.Key, value domain.Value) error
	// Delete removes key. Deleting an absent key is not an error (idempotent).
	Delete(ctx context.Context, key domain.Key) error
}

// StorageEngine is the outbound (driven) port: the contract the core relies on
// to actually persist data.
//
// This is the seam that makes the DDIA journey possible. Today the only
// implementation is an in-memory hash map. As the book progresses we add new
// implementations — an append-only log with a hash index, SSTables + an
// LSM-tree, a B-tree — each as a separate adapter, with no change to the
// domain or the service above it.
//
// context.Context is threaded through so slow operations (a future disk flush
// or compaction) can honour request cancellation and deadlines.
type StorageEngine interface {
	Get(ctx context.Context, key domain.Key) (domain.Value, error)
	Set(ctx context.Context, key domain.Key, value domain.Value) error
	Delete(ctx context.Context, key domain.Key) error
	// Close releases engine resources (files, background workers). In-memory
	// engines make it a no-op, but the port declares it so callers can always
	// shut an engine down cleanly.
	Close() error
}
