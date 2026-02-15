[![CI](https://github.com/pscheid92/uuid/actions/workflows/ci.yml/badge.svg)](https://github.com/pscheid92/uuid/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/pscheid92/uuid.svg)](https://pkg.go.dev/github.com/pscheid92/uuid)
[![Go Report Card](https://goreportcard.com/badge/github.com/pscheid92/uuid?v=2)](https://goreportcard.com/report/github.com/pscheid92/uuid)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

# uuid

A modern, zero-alloc, zero-dependency Go UUID library implementing [RFC 9562](https://www.rfc-editor.org/rfc/rfc9562). Built for Go 1.26+ with first-class support for V7 timestamp-ordered UUIDs, pooled generation, and batch APIs.

```
go get github.com/pscheid92/uuid
```

## Quick Start

```go
import "github.com/pscheid92/uuid"

id := uuid.NewV4()                                       // random UUID
id  = uuid.NewV7()                                       // timestamp-ordered, database-friendly
id, err := uuid.Parse("550e8400-e29b-41d4-a716-446655440000") // parse a string
fmt.Println(id.String())                                  // "550e8400-e29b-41d4-a716-446655440000"
```

## Supported Versions

| Version | Description | Function |
|---------|-------------|----------|
| V3 | Deterministic (MD5) | `NewV3(namespace, name)` |
| V4 | Random | `NewV4()` / `Pool.NewV4()` / `NewV4Batch(n)` |
| V5 | Deterministic (SHA-1) | `NewV5(namespace, name)` |
| V7 | Timestamp + random | `NewV7()` / `Pool.NewV7()` / `Generator.NewV7Batch(n)` |
| V8 | Custom data | `NewV8(data)` |

## Usage

### Generation

```go
// Random (V4) - most common
id := uuid.NewV4()

// Timestamp-ordered (V7) - recommended for new systems, database-friendly
id := uuid.NewV7()

// Deterministic (V5, SHA-1) - same inputs always produce the same UUID
id := uuid.NewV5(uuid.NamespaceDNS, "www.example.com")
// 2ed6657d-e927-568b-95e1-2665a8aea6a2

// Deterministic (V3, MD5) - prefer V5 over V3
id := uuid.NewV3(uuid.NamespaceDNS, "www.example.com")
// 5df41881-3aed-3515-88a7-2f4a814cf09e
```

### Parsing & Formatting

`Parse` is strict - it only accepts the standard 36-character hyphenated form:

```go
id, err := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
```

`ParseLenient` additionally accepts URN, braced, and compact forms:

```go
id, _ := uuid.ParseLenient("urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8")
id, _ := uuid.ParseLenient("{6ba7b810-9dad-11d1-80b4-00c04fd430c8}")
id, _ := uuid.ParseLenient("6ba7b8109dad11d180b400c04fd430c8")
```

`MustParse` panics on failure, useful for package-level constants:

```go
var myID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
```

Format back to strings:

```go
id.String() // "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
id.URN()    // "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"
```

### Serialization

UUID implements `encoding.TextMarshaler`/`TextUnmarshaler` (JSON), `database/sql.Scanner`, and `driver.Valuer` (SQL). Use a `*UUID` pointer for nullable fields:

```go
type User struct {
    ID       uuid.UUID  `json:"id"`
    ParentID *uuid.UUID `json:"parent_id"` // null in JSON, SQL NULL when nil
}

// JSON: {"id":"550e8400-e29b-41d4-a716-446655440000","parent_id":null}

var id uuid.UUID
err := row.Scan(&id)
```

UUIDs are sortable via `uuid.Compare`:

```go
slices.SortFunc(ids, uuid.Compare)
```

## Why This Library?

Go already has [google/uuid](https://github.com/google/uuid) and [gofrs/uuid](https://github.com/gofrs/uuid). Here's what this one does differently:

- **Zero allocations**: NewV4, NewV7, Parse, MarshalText, and UnmarshalText all allocate nothing. Other libraries allocate at least once per call.
- **High-throughput APIs**: Pool (~14x faster V4, ~2x faster V7) and Batch (~25x faster bulk V4) amortize `crypto/rand` cost. No equivalent exists in other libraries.
- **V7 monotonicity built-in**: Sub-millisecond ordering via RFC 9562 Method 3, with automatic counter fallback. No configuration needed.
- **No global mutable state**: No `SetRand`, no global clock. V3/V4/V5/V8 are pure functions. V7 monotonicity is scoped to a `Generator` instance.
- **Strict by default**: `Parse` accepts only `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`. Use `ParseLenient` when you explicitly want URN, braced, or compact forms.
- **Simple value type**: `UUID` is `[16]byte`: comparable, copyable, safe as map key. No `NullUUID` - use `*UUID` for nullable SQL/JSON fields.
- **Modern Go, zero dependencies**: Targets Go 1.26+, uses `crypto/rand` (infallible), `encoding.TextAppender`, `hash.Cloner`. Only stdlib. No legacy baggage, no V1/V2/V6.

## Further Reading

- **[Advanced Usage](docs/advanced.md)**: V7 monotonicity, high-throughput Pool and Batch APIs, properties, namespace constants.
- **[Internals](docs/internals.md)**: V7 bit layout, sub-millisecond precision, monotonic counter fallback, Pool amortization, hash.Cloner optimization, parse lookup table.

## Benchmarks

All generation and formatting hot paths are zero-alloc. Compared to [google/uuid](https://github.com/google/uuid) and [gofrs/uuid](https://github.com/gofrs/uuid) on Apple M2:

| Benchmark | pscheid92/uuid | google/uuid | gofrs/uuid |
|-----------|---------------|-------------|------------|
| NewV4 | **247 ns** | 291 ns | 274 ns |
| NewV4 (Pool) | **17 ns** | - | - |
| NewV4Batch(100) | **1,025 ns** | 25,483 ns | 24,768 ns |
| NewV7 | **106 ns** | 309 ns | 130 ns |
| NewV7 (Pool) | **50 ns** | - | - |
| NewV7Batch(100) | **800 ns** | 30,285 ns | 12,410 ns |
| Parse | **23 ns** | 21 ns | 27 ns |
| MarshalText | **11 ns** | 18 ns | 27 ns |

All entries for this library are zero-alloc; other libraries allocate 1 per operation (generation) or 1 per call (formatting). Run the comparison benchmarks yourself:

```bash
cd bench && go test -bench=. -benchmem ./...
```

## License

MIT License. See [LICENSE](LICENSE) for details.
