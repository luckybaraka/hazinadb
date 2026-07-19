package service

import (
	"context"
	"errors"
	"testing"

	"github.com/luckybaraka/hazinadb/internal/core/domain"
)

// fakeEngine is a stand-in for the StorageEngine port. Because the service
// depends on the port and not a concrete engine, we can test it with this
// in-test double — no map, no HTTP, no disk.
type fakeEngine struct {
	setCalls int
	store    map[domain.Key]domain.Value
}

func newFakeEngine() *fakeEngine {
	return &fakeEngine{store: map[domain.Key]domain.Value{}}
}

func (f *fakeEngine) Get(_ context.Context, key domain.Key) (domain.Value, error) {
	v, ok := f.store[key]
	if !ok {
		return nil, domain.ErrKeyNotFound
	}
	return v, nil
}

func (f *fakeEngine) Set(_ context.Context, key domain.Key, value domain.Value) error {
	f.setCalls++
	f.store[key] = value
	return nil
}

func (f *fakeEngine) Delete(_ context.Context, key domain.Key) error {
	delete(f.store, key)
	return nil
}

func (f *fakeEngine) Close() error { return nil }

func TestKeyValue_PutRejectsEmptyKeyBeforeTouchingEngine(t *testing.T) {
	engine := newFakeEngine()
	svc := NewKeyValue(engine)

	err := svc.Put(context.Background(), "", domain.Value("x"))
	if !errors.Is(err, domain.ErrEmptyKey) {
		t.Fatalf("got %v, want ErrEmptyKey", err)
	}
	if engine.setCalls != 0 {
		t.Fatalf("engine was called %d times; validation should short-circuit", engine.setCalls)
	}
}

func TestKeyValue_PutThenGet(t *testing.T) {
	svc := NewKeyValue(newFakeEngine())
	ctx := context.Background()

	if err := svc.Put(ctx, "k", domain.Value("v")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := svc.Get(ctx, "k")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != "v" {
		t.Fatalf("got %q, want %q", got, "v")
	}
}
