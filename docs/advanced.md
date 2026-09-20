# Advanced Usage

This document covers features beyond basic UUID generation and parsing. For getting started, see the [README](../README.md).

## V7 Monotonicity

Multiple V7 UUIDs generated within the same millisecond are guaranteed to sort in creation order within a single `Generator`:

```go
gen := uuid.NewGenerator()
id1 := gen.NewV7()
id2 := gen.NewV7() // guaranteed id1 < id2 even within the same millisecond
```

The package-level `uuid.NewV7()` uses a default shared generator, so it also provides monotonicity out of the box. Create a dedicated `Generator` when you need isolated monotonicity guarantees (e.g., per-request or per-goroutine ordering).

See [Internals: V7 Monotonic Counter Fallback](internals.md#v7-monotonic-counter-fallback) for how this works under the hood.

## Backfilling with NewV7At

When a table gets V7 keys after rows already exist, `NewV7At` builds a V7 UUID for a past creation time so the backfilled keys sort among live ones at the right position:

```go
id := uuid.NewV7At(row.CreatedAt)
```

The millisecond field and the 12-bit sub-millisecond fraction come from the given time, laid out exactly as the live generator lays them out, and the remaining 62 bits are random. Two calls with the same time tie on their first 8 bytes and are ordered only by the random tail.

`NewV7At` is a pure function and never touches the monotonic state of any `Generator` or `Pool`, so backfilling with a future timestamp cannot push live UUIDs ahead of the wall clock. It panics for times outside the 48-bit range (before 1970 or after roughly the year 10889); a zero `time.Time` is out of range, so an uninitialized field fails loudly instead of producing a silently wrong key.

## High-Throughput Generation

For hot paths, `Pool` amortizes the cost of `crypto/rand` by pre-generating random bytes in bulk:

```go
pool := uuid.NewPool()
id := pool.NewV4() // ~14x faster than NewV4()
id  = pool.NewV7() // ~2x faster than NewV7() (time.Now dominates)
```

For bulk workloads (database seeding, ETL, load testing), batch APIs generate many UUIDs with a single `crypto/rand` call:

```go
ids := uuid.NewV4Batch(1000) // ~25x faster than calling NewV4() in a loop
ids  = uuid.NewV7Batch(1000) // ~13x faster, all monotonically increasing
```

`uuid.NewV7Batch` uses the package-level default generator; call `NewV7Batch` on a dedicated `Generator` for isolated monotonicity guarantees.

Both `Pool` and `Batch` draw exclusively from `crypto/rand`, and `Pool` is safe for concurrent use. Each `Pool` keeps its own V7 monotonic state, independent of the package-level `NewV7` generator and of any other `Pool` or `Generator`; UUIDs drawn from different sources are not ordered relative to each other, so pick one source per ordering domain. One caveat: `Pool` buffers pre-generated randomness in process memory, so it is not fork-safe — a forked process or a cloned/restored VM snapshot duplicates the buffer and can emit identical UUIDs from both copies. Use the package-level functions where that matters. The batch APIs are unaffected since they read fresh randomness on every call.

See [Internals: Pool](internals.md#pool-amortizing-cryptorand) for how pooling works.

## Properties

```go
id := uuid.NewV7()

id.Version()  // uuid.V7
id.Variant()  // uuid.VariantRFC9562
id.IsNil()    // false
id.Time()     // (time.Time, bool): millisecond precision; ok is false for non-V7
id.Bytes()    // []byte copy of the 16 raw bytes
```

`Compare(a, b UUID) int` returns -1, 0, or +1 for use with `slices.SortFunc`:

```go
slices.SortFunc(ids, uuid.Compare)
```

## SQL: BINARY(16) Columns

`Value` returns the 36-character string, which is what native `uuid` column types (PostgreSQL, SQLite, CockroachDB) expect. MySQL and MariaDB have no native type and conventionally store UUIDs in `BINARY(16)`. `Scan` already accepts 16 raw bytes, so only `Value` needs to change. Wrap the type:

```go
type BinaryUUID struct{ uuid.UUID }

func (b BinaryUUID) Value() (driver.Value, error) {
	return b.Bytes(), nil
}
```

`Scan`, `String`, `MarshalText`, and every other method are promoted from the embedded `uuid.UUID`, so the wrapper behaves identically everywhere except when written to the database.

## Namespace Constants

Predefined namespace UUIDs for use with `NewV5` ([RFC 9562 Appendix C](https://www.rfc-editor.org/rfc/rfc9562#appendix-C)):

```go
uuid.NamespaceDNS   // 6ba7b810-9dad-11d1-80b4-00c04fd430c8
uuid.NamespaceURL   // 6ba7b811-9dad-11d1-80b4-00c04fd430c8
uuid.NamespaceOID   // 6ba7b812-9dad-11d1-80b4-00c04fd430c8
uuid.NamespaceX500  // 6ba7b814-9dad-11d1-80b4-00c04fd430c8
```

See [pkg.go.dev](https://pkg.go.dev/github.com/pscheid92/uuid) for the full API reference.
