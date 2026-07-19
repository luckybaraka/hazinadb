# HazinaDB

A storage engine built from scratch, one chapter of *Designing Data-Intensive
Applications* (DDIA) at a time. It starts as a simple key-value store and grows
into real storage-engine techniques — hash indexes, SSTables/LSM-trees, B-trees.

> **hazina** — Swahili for *treasure / store*.

## Current status

A concurrency-safe, in-memory key-value store exposed over HTTP. Data lives in
RAM and is lost on restart — durability arrives in later chapters.

## Quick start

```bash
# Run the server (defaults to :8000)
go run ./cmd/hazinadb
# or choose an address
go run ./cmd/hazinadb --addr :9000
```

```bash
# Store a value (the request body is the value)
curl -X PUT   localhost:8000/kv/greeting -d 'hello world'

# Read it back
curl          localhost:8000/kv/greeting      # -> hello world

# Delete it
curl -X DELETE localhost:8000/kv/greeting

# Missing keys return 404
curl -i       localhost:8000/kv/nope           # -> 404 key not found
```

| Method   | Path         | Body  | Success        |
|----------|--------------|-------|----------------|
| `PUT`    | `/kv/{key}`  | value | `204 No Content` |
| `GET`    | `/kv/{key}`  | —     | `200` + value  |
| `DELETE` | `/kv/{key}`  | —     | `204 No Content` |

## Architecture — Hexagonal (Ports & Adapters)

The **core** (domain + application service) is the center of the hexagon and
knows nothing about HTTP or maps. The outside world plugs in through **ports**
(interfaces), implemented by **adapters**.

```
   driving side                CORE (the hexagon)                driven side
  (who calls us)                                              (what we depend on)

  ┌──────────────┐   inbound port    ┌───────────────┐  outbound port  ┌──────────────┐
  │  REST / HTTP │ ───────────────▶  │  application  │ ───────────────▶ │  in-memory   │
  │   adapter    │  KeyValueStore    │   service     │  StorageEngine   │  map engine  │
  └──────────────┘                   │      │        │                  └──────────────┘
                                     │      ▼        │                  (swap for LSM,
                                     │   domain      │                   B-tree, etc.)
                                     │ (Key, Value)  │
                                     └───────────────┘

  Dependencies always point INWARD. The core never imports an adapter.
```

**Why this shape for a storage engine?** DDIA is a tour of *different storage
engines behind the same interface*. `StorageEngine` is that interface. Each new
engine is a new adapter under `internal/adapter/outbound/` that the core and the
HTTP layer never notice. `main.go` is the single place that picks one.

## Layout

```
cmd/hazinadb/            Composition root — wires adapters to the core, starts the server
internal/
  core/                  The hexagon. Depends on nothing outside itself.
    domain/              Business concepts: Key, Value, validation, domain errors
    port/                The boundaries: inbound + outbound interfaces
    service/             Application logic implementing the inbound port
  adapter/
    inbound/  rest/      Driving adapter: HTTP transport
    outbound/ memory/    Driven adapter: in-memory storage engine
```

Rule of thumb: **`core/` may never import `adapter/`.** If you feel the urge to,
the missing piece is a new port.

## Development

```bash
make run     # run the server
make test    # run all tests
make check   # fmt + vet + test  (run before committing)
```

## Roadmap (following DDIA, Chapter 3)

- [x] In-memory hash-map engine
- [ ] Append-only log with an in-memory hash index
- [ ] Log segmentation + compaction
- [ ] SSTables and an LSM-tree engine
- [ ] B-tree engine
- [ ] Write-ahead log for crash recovery

Each item is a new outbound adapter satisfying `port.StorageEngine`.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for the deeper "why".
