[![CI](https://github.com/pscheid92/uuid/actions/workflows/ci.yml/badge.svg)](https://github.com/pscheid92/uuid/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/pscheid92/uuid.svg)](https://pkg.go.dev/github.com/pscheid92/uuid)
[![Go Report Card](https://goreportcard.com/badge/github.com/pscheid92/uuid?v=2)](https://goreportcard.com/report/github.com/pscheid92/uuid)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

# uuid

A companion to Go's standard library [`uuid`](https://pkg.go.dev/uuid) package. The standard library (Go 1.27+) generates V4 and V7 UUIDs and parses and prints them. This package adds what it leaves out: V5 and V8, version and timestamp accessors, strict parsing with typed errors, V7 backfilling and per-instance ordering, and bulk generation. It implements [RFC 9562](https://www.rfc-editor.org/rfc/rfc9562) and has no dependencies.

Both packages define `UUID` as `[16]byte`, so a value converts between them at no cost and each can be used where it fits.

```
go get github.com/pscheid92/uuid
```

## Do You Need It?

If you only create random or time-ordered IDs and print or parse them, the standard library is enough, and you should use it. Add this package when you need something in the right-hand column:

| You need to | Standard library `uuid` | This package |
|-------------|-------------------------|--------------|
| Derive the same ID from the same name | - | `NewV5`, `NewV5Bytes` |
| Build a UUID from your own bits | - | `NewV8` |
| Inspect a UUID | - | `Version`, `Variant`, `Time` (V7 timestamp) |
| Reject non-canonical input | `Parse` accepts four forms | `Parse` accepts only the 36-character form; `ParseLenient` accepts all four |
| Tell bad input from other errors | a plain `invalid uuid` error | `ErrInvalid`, `ParseError` with the input and position |
| Give existing rows V7 keys that sort by creation time | - | `NewV7At` |
| Keep separate V7 orderings | one sequence per process | `Generator`, one sequence each |
| Keep V7 order when the clock steps back | not guaranteed | guaranteed within a `Generator` |
| Generate many UUIDs at once | - | `NewV4Batch`, `NewV7Batch`, `FillV4`, `FillV7`, `Pool` |
| Encode as raw bytes through `encoding.BinaryMarshaler` (gob and similar) | `u[:]` only | `MarshalBinary`, `UnmarshalBinary`, `AppendBinary`, `FromBytes` |

Two things are not on the list. Since Go 1.27, `database/sql` scans into the standard library's `UUID` itself and its default converter binds one as a string, so SQL support is not a reason to switch; this package's `Scan` and `Value` do the same job for its own type. And speed: single UUIDs cost the same in both packages; see [Performance](#performance).

## Using It with the Standard Library

This package also has `NewV4`, `NewV7`, `Parse`, `MustParse`, and `String`, so code that needs its extras can use it alone. Where a dependency hands you a standard library UUID, or expects one, convert:

```go
import (
    stdlib "uuid"

    "github.com/pscheid92/uuid"
)

id := uuid.UUID(stdlib.New())      // stdlib -> this package: gain Version, Time, ...
std := stdlib.UUID(uuid.NewV7())   // this package -> stdlib
```

The conversion is free: both types are `[16]byte`. UUIDs from the standard library's `NewV7` and from this package's are not strictly ordered relative to each other, so pick one source for keys that must sort.

> [!IMPORTANT]
> The two packages share the name `uuid` and many function names (`NewV7`, `Parse`, `MustParse`, ...). When goimports (or an editor using its import logic, such as gopls) adds a missing import for `uuid.NewV7()`, it picks the standard library's `"uuid"` unless other files in the same package already import this one and use every `uuid` name the new file needs. So in a new package, or for the first call of a function no other file uses yet, you get the standard library. That file may still compile, but its UUIDs are the standard library's type: its `Parse` is lenient, and `Version`, `Time`, and the rest of this package's API are missing. Check that the import reads `github.com/pscheid92/uuid`, and alias the standard library package (`stdlib "uuid"`) in files that need both.

## Quick Start

```go
import "github.com/pscheid92/uuid"

id := uuid.NewV7()                                             // timestamp-ordered key
old := uuid.NewV7At(row.CreatedAt)                             // V7 key for a row that already exists
ts, ok := id.Time()                                            // when a V7 UUID was created
tenant := uuid.NewV5(uuid.NamespaceDNS, "customer.example")    // same name, same UUID
id, err := uuid.Parse(r.PathValue("id"))                       // strict: canonical form only
if errors.Is(err, uuid.ErrInvalid) {
    http.Error(w, "invalid id", http.StatusBadRequest)
}
```

## Supported Versions

| Version | Description | Function |
|---------|-------------|----------|
| V4 | Random | `NewV4()` / `Pool.NewV4()` / `NewV4Batch(n)` / `FillV4(dst)` |
| V5 | Deterministic (SHA-1) | `NewV5(namespace, name)` / `NewV5Bytes(namespace, name)` |
| V7 | Timestamp + random | `NewV7()` / `NewV7At(t)` / `Pool.NewV7()` / `NewV7Batch(n)` / `FillV7(dst)` |
| V8 | Custom data | `NewV8(data)` |

## Usage

### Generation

```go
// Random (V4) - most common
id := uuid.NewV4()

// Timestamp-ordered (V7) - recommended for new systems, database-friendly
id := uuid.NewV7()

// Timestamp-ordered at a given time (V7) - backfill existing rows so they sort correctly
id := uuid.NewV7At(row.CreatedAt)

// Deterministic (V5, SHA-1) - same inputs always produce the same UUID
id := uuid.NewV5(uuid.NamespaceDNS, "www.example.com")
// 2ed6657d-e927-568b-95e1-2665a8aea6a2
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

Which form is accepted depends on where the text comes from:

| Entry point | Accepts |
|-------------|---------|
| `Parse`, `MustParse`, `UnmarshalText` (JSON, XML, and other text encodings) | Only the 36-character form |
| `ParseLenient`, `Scan` (`database/sql`) | All four forms |

Text encodings are strict because they are usually API input, where anything but the canonical form is a client bug worth rejecting. `Scan` is lenient because databases and drivers return UUIDs in different forms. google/uuid and the standard library parse leniently everywhere, so code moving from them should call `ParseLenient` where it relied on that. To accept every form in JSON, wrap the type (see the `Example (LenientJSON)` in the [package docs](https://pkg.go.dev/github.com/pscheid92/uuid)):

```go
type LenientUUID struct{ uuid.UUID }

func (l *LenientUUID) UnmarshalText(b []byte) (err error) {
	l.UUID, err = uuid.ParseLenient(string(b))
	return err
}
```

`MustParse` panics on failure, useful for package-level constants:

```go
var myID = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
```

Every error this package returns for malformed input matches `uuid.ErrInvalid`: from parsing, text and binary decoding (including a malformed UUID string inside JSON), and `Scan`. Errors that `encoding/json` itself raises, such as a syntax error or a number where a string belongs, do not. Use `errors.AsType[*uuid.ParseError]` when you need the offending input:

```go
if errors.Is(err, uuid.ErrInvalid) {
    http.Error(w, "invalid id", http.StatusBadRequest)
}
```

Format back to strings:

```go
id.String() // "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
id.URN()    // "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"

fmt.Printf("%v", id)    // 6ba7b810-9dad-11d1-80b4-00c04fd430c8
fmt.Printf("%x", id[:]) // 6ba7b8109dad11d180b400c04fd430c8 (the 16 bytes as hex)
```

`%x` on the UUID itself hex-encodes the 36-character string (`fmt` uses `String` for it), so format `id[:]` when you want the 32 hex digits.

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

UUIDs are sortable via `uuid.Compare` (or the `Compare` method, as in the standard library):

```go
slices.SortFunc(ids, uuid.Compare)
```

## Design

- **Strict by default**: `Parse` and text decoding accept only `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`, because anything else in API input is usually a client bug. `ParseLenient` and `Scan` accept URN, braced, and compact forms.
- **V7 ordering you can reason about**: Sub-millisecond ordering via RFC 9562 Method 3, with a counter that keeps UUIDs increasing when calls outrun the clock or the clock steps back. The ordering lives in a `Generator`: the package-level `NewV7` uses a shared default one (like `http.DefaultClient`), and you can create your own for isolated ordering.
- **No global configuration**: No `SetRand`, no swappable clock or random source. Randomness always comes from `crypto/rand`. V4/V5/V8 are stateless.
- **Simple value type**: `UUID` is `[16]byte`, the same as the standard library's: comparable, copyable, safe as map key. No `NullUUID` - use `*UUID` for nullable SQL/JSON fields.
- **No allocations where avoidable**: generation, parsing, and `AppendText` allocate nothing; `FillV4`/`FillV7` reuse a buffer.
- **Only the standard library underneath**: zero dependencies, Go 1.27+. V1/V2/V3/V6 are parsed and identified, never generated.

## Further Reading

- **[Advanced Usage](docs/advanced.md)**: V7 monotonicity, high-throughput Pool and Batch APIs, properties, namespace constants.
- **[Internals](docs/internals.md)**: V7 bit layout, sub-millisecond precision, monotonic counter fallback, Pool amortization, zero-alloc V5 hashing, parse lookup table.

## Performance

Performance is not the reason to choose this package over the standard library.

- **Single UUIDs** cost the same in both, since both spend their time in the same `crypto/rand` read.
- **Encoding** (String, MarshalText) takes about 20% less time here on Linux and 30–35% less on macOS, which is 5–10 ns per call.
- **Parsing** is 25–30% faster here on Apple silicon and 5–10% slower on the x86-64 runners; either way a few nanoseconds.
- **Batches** are the exception. A batch of 100 is ~2x faster than 100 single V4 calls on Linux and ~5x for V7. On macOS, where the 16-byte `crypto/rand` read behind a V4 UUID costs three to four times as much, it is ~30x and ~15x.
- **`Pool`** gains ~1.6–2x for V4 and ~1.1–1.2x for V7 on Linux (~14x and ~2x on macOS), so measure on the hardware you deploy to before adding one.

The tables also cover [google/uuid](https://github.com/google/uuid) and [gofrs/uuid](https://github.com/gofrs/uuid), for code moving from them. google/uuid can pool V4 and V7 randomness behind a process-wide toggle (`EnableRandPool`, not safe to flip while generating); no other library generates in batches.

Compared to the Go 1.27 standard library `uuid` package, [google/uuid](https://github.com/google/uuid), and [gofrs/uuid](https://github.com/gofrs/uuid) on Apple M2 (fastest of 8–12 runs; the fastest entry per row is bold):

| Benchmark | pscheid92/uuid | stdlib (Go 1.27) | google/uuid | gofrs/uuid |
|-----------|---------------|------------------|-------------|------------|
| NewV4 | 240 ns | **237 ns** | 248 ns | 243 ns |
| NewV4 (Pool) | **16 ns** | - | 29 ns¹ | - |
| NewV4Batch(100) | **752 ns** | - | 24,910 ns² | 24,538 ns² |
| NewV5 | 63 ns | - | 100 ns | **62 ns** |
| NewV7 | 104 ns | **101 ns** | 296 ns | 117 ns |
| NewV7 (Pool) | **49 ns** | - | 73 ns¹ | - |
| NewV7Batch(100) | **691 ns** | - | 29,768 ns² | 11,615 ns² |
| Parse | **18 ns** | 25 ns | 19 ns | 27 ns |
| UnmarshalText | **18 ns** | 25 ns | 19 ns | 27 ns |
| String | **20 ns** | 29 ns | 26 ns | 24 ns |
| MarshalText | **17 ns** | 27 ns | 24 ns | 21 ns |

The same benchmarks on Linux (GitHub Actions `ubuntu-latest`, AMD EPYC 9V45, 4 vCPUs; median of 6 runs from the CI `bench` job, spread mostly within ±5%; runner CPUs vary, and the ratios between rows and columns hold better than the absolute numbers):

| Benchmark | pscheid92/uuid | stdlib (Go 1.27) | google/uuid | gofrs/uuid |
|-----------|---------------|------------------|-------------|------------|
| NewV4 | **72 ns** | **73 ns** | 85 ns | 82 ns |
| NewV4 (Pool) | **46 ns** | - | 62 ns¹ | - |
| NewV4Batch(100) | **3,757 ns** | - | 8,377 ns² | 7,968 ns² |
| NewV5 | **78 ns** | - | 115 ns | **77 ns** |
| NewV7 | **101 ns** | 103 ns | 149 ns | 116 ns |
| NewV7 (Pool) | **91 ns** | - | 130 ns¹ | - |
| NewV7Batch(100) | **2,166 ns** | - | 14,850 ns² | 11,440 ns² |
| Parse | 19 ns | **17 ns** | 18 ns | 21 ns |
| UnmarshalText | 17 ns | **15 ns** | 17 ns | 20 ns |
| String | **22 ns** | 26 ns | 24 ns | 27 ns |
| MarshalText | **19 ns** | 24 ns | 22 ns | 23 ns |

¹ With `google.EnableRandPool()`, which also removes the allocation. ² No batch API; the benchmark makes 100 single calls.

All entries for this library and the standard library are zero-alloc except the batches, which allocate their result, and String and MarshalText, which must return a newly allocated result in every library; use `AppendText` to encode into your own buffer without allocating. google/uuid allocates on every generation call, except V4 and V7 with its pool enabled; gofrs/uuid allocates on every generation call except NewV5. Run the comparison benchmarks yourself:

```bash
cd bench && go test -bench=. -benchmem ./...
```

## License

MIT License. See [LICENSE](LICENSE) for details.
