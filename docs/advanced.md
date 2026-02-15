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

gen := uuid.NewGenerator()
ids  = gen.NewV7Batch(1000)  // ~15x faster, all monotonically increasing
```

Both `Pool` and `Batch` use `crypto/rand` exclusively - no security trade-offs. `Pool` is safe for concurrent use.

See [Internals: Pool](internals.md#pool-amortizing-cryptorand) for how pooling works.

## Properties

```go
id := uuid.NewV7()

id.Version()  // uuid.Version7
id.Variant()  // uuid.VariantRFC9562
id.IsNil()    // false
id.Time()     // time.Time (millisecond precision, V7 only)
id.Bytes()    // [16]byte
```

`Compare(a, b UUID) int` returns -1, 0, or +1 for use with `slices.SortFunc`:

```go
slices.SortFunc(ids, uuid.Compare)
```

## Namespace Constants

Predefined namespace UUIDs for use with `NewV3` and `NewV5` ([RFC 9562 Appendix C](https://www.rfc-editor.org/rfc/rfc9562#appendix-C)):

```go
uuid.NamespaceDNS   // 6ba7b810-9dad-11d1-80b4-00c04fd430c8
uuid.NamespaceURL   // 6ba7b811-9dad-11d1-80b4-00c04fd430c8
uuid.NamespaceOID   // 6ba7b812-9dad-11d1-80b4-00c04fd430c8
uuid.NamespaceX500  // 6ba7b814-9dad-11d1-80b4-00c04fd430c8
```

See [pkg.go.dev](https://pkg.go.dev/github.com/pscheid92/uuid) for the full API reference.
