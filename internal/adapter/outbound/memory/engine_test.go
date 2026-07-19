package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/luckybaraka/hazinadb/internal/core/domain"
)

func TestEngine_SetThenGet(t *testing.T) {
	e := NewEngine()
	ctx := context.Background()

	if err := e.Set(ctx, "greeting", domain.Value("hello")); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := e.Get(ctx, "greeting")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}

func TestEngine_GetMissingKey(t *testing.T) {
	e := NewEngine()

	_, err := e.Get(context.Background(), "absent")
	if !errors.Is(err, domain.ErrKeyNotFound) {
		t.Fatalf("got %v, want ErrKeyNotFound", err)
	}
}

func TestEngine_Delete(t *testing.T) {
	e := NewEngine()
	ctx := context.Background()

	_ = e.Set(ctx, "temp", domain.Value("x"))
	if err := e.Delete(ctx, "temp"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := e.Get(ctx, "temp"); !errors.Is(err, domain.ErrKeyNotFound) {
		t.Fatalf("key still present after delete")
	}

	// Deleting an absent key is a no-op, not an error.
	if err := e.Delete(ctx, "never-existed"); err != nil {
		t.Fatalf("delete of absent key should be a no-op, got %v", err)
	}
}

// TestEngine_StoredValueIsIsolated proves the copy-on-write/read guarantee:
// mutating the caller's buffer must not corrupt the stored value.
func TestEngine_StoredValueIsIsolated(t *testing.T) {
	e := NewEngine()
	ctx := context.Background()

	buf := domain.Value("original")
	_ = e.Set(ctx, "k", buf)
	buf[0] = 'X' // caller mutates its buffer after Set

	got, _ := e.Get(ctx, "k")
	if string(got) != "original" {
		t.Fatalf("stored value was mutated through caller's buffer: %q", got)
	}
}
