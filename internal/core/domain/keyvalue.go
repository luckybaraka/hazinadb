// Package domain holds the pure business concepts of HazinaDB.
//
// It has ZERO dependencies on transport (HTTP), storage engines (maps, disk),
// or any framework. Everything else in the codebase points inward at this
// package — never the other way around. This is the center of the hexagon.
package domain

import "errors"

// Key identifies a single record. Modelled as a distinct type (not a bare
// string) so validation rules live in one place and the compiler stops us
// from accidentally passing an arbitrary string where a key is expected.
type Key string

// Value is the opaque payload stored under a Key.
//
// We use []byte rather than string on purpose: the storage engine treats
// values as raw bytes (that's how they'll eventually be written to disk),
// and HTTP bodies already arrive as []byte. Keeping []byte end-to-end avoids
// a string<->[]byte copy on every read and write.
type Value []byte

// Domain errors. These are sentinel values, so callers compare them with
// errors.Is(err, domain.ErrKeyNotFound) — which works reliably, unlike
// comparing freshly-allocated error pointers.
var (
	// ErrKeyNotFound means no record exists for the requested key.
	ErrKeyNotFound = errors.New("key not found")
	// ErrEmptyKey means the caller supplied a key with no bytes.
	ErrEmptyKey = errors.New("key must not be empty")
)

// Validate enforces the invariants every key must satisfy before it reaches
// the storage engine. Returning early here keeps invalid data out of the core.
func (k Key) Validate() error {
	if len(k) == 0 {
		return ErrEmptyKey
	}
	return nil
}
