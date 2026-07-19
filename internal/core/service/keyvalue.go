// Package service contains the application logic of the core: it implements the
// inbound port by orchestrating domain rules and delegating persistence to the
// outbound StorageEngine port.
//
// Note what this package does NOT import: no net/http, no map, no files. It
// depends only on domain + port. That is what keeps the core testable in
// isolation and independent of any technology choice.
package service

import (
	"context"

	"github.com/luckybaraka/hazinadb/internal/core/domain"
	"github.com/luckybaraka/hazinadb/internal/core/port"
)

// KeyValue implements the inbound port.KeyValueStore. This compile-time
// assertion fails the build the moment the implementation drifts from the port.
var _ port.KeyValueStore = (*KeyValue)(nil)

// KeyValue is the application service. It holds the engine behind its outbound
// port, so it can be wired to any StorageEngine implementation.
type KeyValue struct {
	engine port.StorageEngine
}

// NewKeyValue wires the service to a storage engine. The engine is injected
// (dependency inversion), so main.go decides which concrete engine is used.
func NewKeyValue(engine port.StorageEngine) *KeyValue {
	return &KeyValue{engine: engine}
}

// Get validates the key, then reads through to the engine.
func (s *KeyValue) Get(ctx context.Context, key domain.Key) (domain.Value, error) {
	if err := key.Validate(); err != nil {
		return nil, err
	}
	return s.engine.Get(ctx, key)
}

// Put validates the key, then writes through to the engine.
func (s *KeyValue) Put(ctx context.Context, key domain.Key, value domain.Value) error {
	if err := key.Validate(); err != nil {
		return err
	}
	return s.engine.Set(ctx, key, value)
}

// Delete validates the key, then removes it via the engine.
func (s *KeyValue) Delete(ctx context.Context, key domain.Key) error {
	if err := key.Validate(); err != nil {
		return err
	}
	return s.engine.Delete(ctx, key)
}
