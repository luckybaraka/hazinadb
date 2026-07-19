// Package memory is a driven (outbound) adapter: an in-memory implementation of
// the port.StorageEngine contract backed by a Go map.
//
// It is the simplest possible engine — the DDIA starting point. Data lives only
// in RAM and is lost on restart. Later chapters replace it with durable engines
// (append-only log, LSM-tree, B-tree) that satisfy the same port.
package memory

import (
	"context"
	"sync"

	"github.com/luckybaraka/hazinadb/internal/core/domain"
	"github.com/luckybaraka/hazinadb/internal/core/port"
)

// Engine implements port.StorageEngine using a hash map. The assertion keeps
// the implementation honest against the port.
var _ port.StorageEngine = (*Engine)(nil)

// Engine is a concurrency-safe in-memory key-value store.
type Engine struct {
	// A RWMutex (not a plain Mutex) lets any number of readers proceed in
	// parallel and only serialises writers. For a read-heavy store this is a
	// meaningful throughput win over a Mutex, at no extra complexity.
	mu   sync.RWMutex
	data map[domain.Key]domain.Value
}

// NewEngine returns a ready-to-use in-memory engine.
func NewEngine() *Engine {
	return &Engine{data: make(map[domain.Key]domain.Value)}
}

// Get returns a copy of the stored value, or domain.ErrKeyNotFound.
func (e *Engine) Get(_ context.Context, key domain.Key) (domain.Value, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	value, ok := e.data[key]
	if !ok {
		return nil, domain.ErrKeyNotFound
	}
	// Return a copy. The map holds a []byte; handing out the same slice would
	// let a caller mutate our stored bytes from the outside. Copying keeps the
	// store's data immutable to callers — cheap insurance against action-at-a-
	// distance bugs.
	return append(domain.Value(nil), value...), nil
}

// Set stores a copy of value under key, overwriting any existing entry.
func (e *Engine) Set(_ context.Context, key domain.Key, value domain.Value) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Copy on write for the same reason as Get: the caller may reuse or mutate
	// its buffer after this call returns, so we take our own snapshot.
	e.data[key] = append(domain.Value(nil), value...)
	return nil
}

// Delete removes key. Removing an absent key is a no-op, keeping deletes
// idempotent.
func (e *Engine) Delete(_ context.Context, key domain.Key) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.data, key)
	return nil
}

// Close releases resources. An in-memory engine holds none, so this is a no-op
// that exists only to satisfy the port.
func (e *Engine) Close() error { return nil }
