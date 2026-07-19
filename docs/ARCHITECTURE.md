# Architecture

This document explains the *why* behind HazinaDB's structure. If you only want
to run it, the [README](../README.md) is enough.

## The one rule

**Dependencies point inward.** `core/` never imports `adapter/`. The core
defines interfaces (ports); adapters implement them. Everything else follows
from this.

```
adapter  ──depends on──▶  core   (allowed)
core     ──depends on──▶  adapter (FORBIDDEN — introduce a port instead)
```

## The three layers of the core

### `domain/` — what the business is
Pure types and rules: `Key`, `Value`, their validation, and the sentinel errors
(`ErrKeyNotFound`, `ErrEmptyKey`). No imports beyond the standard library. If you
can't unit-test it without a mock, it probably doesn't belong here.

`Key` and `Value` are distinct types rather than bare `string`/`[]byte`. That
gives validation a home and lets the compiler catch mix-ups.

### `port/` — the boundaries
Two kinds of interface:

- **Inbound (driving) port — `KeyValueStore`.** What the core offers. Driving
  adapters call it. If we add a gRPC or CLI front end, it depends on this and
  nothing deeper.
- **Outbound (driven) port — `StorageEngine`.** What the core needs from
  infrastructure. The service calls it; engines implement it. This is the seam
  the whole DDIA journey pivots on.

### `service/` — the application logic
Implements the inbound port by orchestrating domain rules and delegating
persistence to the outbound port. It validates keys before any I/O, so invalid
data never reaches an engine. It imports `domain` and `port` — never an adapter.

## Adapters

### Driving (inbound) — `adapter/inbound/rest`
Translates HTTP to port calls and domain errors to status codes, and nothing
more. Business rules live in the core, so this stays thin. Swapping in a
different transport means adding a sibling adapter, not editing the core.

### Driven (outbound) — `adapter/outbound/memory`
Implements `StorageEngine` with a Go map guarded by a `sync.RWMutex`. This is
the DDIA starting point: an in-memory hash index. Future engines (log-structured,
LSM, B-tree) are new packages under `adapter/outbound/` that satisfy the same
port.

## Composition root — `cmd/hazinadb`

The only place allowed to name concrete types. It constructs the engine, injects
it into the service, hands the service to the HTTP server, and starts it. To
change which engine runs, edit these few lines — nothing else.

```go
engine := memory.NewEngine()          // pick a driven adapter
store  := service.NewKeyValue(engine) // inject it into the core
server := rest.NewServer(addr, store) // pick a driving adapter
```

## Design decisions worth knowing

- **`Value` is `[]byte`, not `string`.** Values are raw bytes headed for disk,
  and HTTP bodies already arrive as `[]byte`. Staying in `[]byte` end-to-end
  avoids a copy on every read and write.
- **Copy-on-write / copy-on-read in the memory engine.** A map stores a `[]byte`
  by reference; handing the same slice to a caller (or storing a caller's slice
  directly) would let outside code mutate our data. We copy at the boundary. The
  cost is one allocation; the benefit is that stored data is immutable to
  callers. A durable engine serialises to disk anyway, so this cost disappears
  later.
- **`RWMutex`, not `Mutex`.** Reads run in parallel; only writes serialise. A win
  for a read-heavy store, at no added complexity.
- **`context.Context` on the engine port.** A no-op for the in-memory engine, but
  future disk flushes and compactions can honour cancellation and deadlines
  without a signature change.
- **Sentinel errors + `errors.Is`.** Comparing freshly-allocated error pointers
  never matches. Package-level sentinels (`domain.ErrKeyNotFound`) compare
  reliably, which is how the HTTP layer maps a miss to `404`.
- **Graceful shutdown.** On `Ctrl+C`/`SIGTERM` the server drains in-flight
  requests within a timeout before exiting — the habit to keep once writes
  become durable.

## Adding the next engine (worked example)

1. Create `internal/adapter/outbound/log/engine.go`.
2. Implement `port.StorageEngine` (`Get`/`Set`/`Delete`/`Close`).
3. Add `var _ port.StorageEngine = (*Engine)(nil)` so the compiler enforces the
   contract.
4. In `cmd/hazinadb/main.go`, swap `memory.NewEngine()` for `log.NewEngine(...)`.

No change to `domain`, `port`, `service`, or the HTTP adapter. That is the whole
point.
