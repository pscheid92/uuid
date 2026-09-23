# UUID Library

Modern Go UUID library (RFC 9562) targeting Go 1.27. Module: `github.com/pscheid92/uuid`

## Commands

```bash
go test ./...                           # run all tests
go test -race ./...                     # run with race detector
go vet ./...                            # static analysis
go test -bench=. -benchmem ./...        # benchmarks with alloc stats
go test -fuzz=FuzzDecodersAgree -fuzztime=30s ./...   # fuzz all decoders: agreement + round-trip
cd bench && go test -bench=. -benchmem ./...          # comparison benchmarks vs stdlib uuid, google/uuid, gofrs/uuid
```

## Architecture

Single flat package at the module root. Each file has a focused responsibility:

- `uuid.go` — package doc, UUID type, Nil/Max, Namespace constants, Version/Variant types (VNil/V4/V5/V7/V8/VMax), accessors (Version/Variant/IsNil/Bytes/Time), Compare (method and package func)
- `parse.go` — Parse (strict 36-char), ParseLenient (URN/braced/compact), MustParse, FromBytes; hex lookup table + offset array; ParseError, LengthError, and the ErrInvalid sentinel both match via Is
- `format.go` — String, URN, encodeHex, AppendText/Binary, Marshal/Unmarshal (Text + Binary); Scan (database/sql.Scanner), Value (driver.Valuer)
- `generate.go` — NewV4/V5/V7/V8, NewV5Bytes, NewV4Batch/FillV4, NewV7Batch/FillV7 (package-level, use default generator), Generator type with per-instance V7 monotonicity (RFC 9562 Method 3) and NewV7Batch, Pool type with buffered NewV4/NewV7 (zero value ready to use), stack-buffer SHA-1 for V5. Shared inlinable helpers: `v7Seq` (clock → ordering value), `reserve` (monotonic step), `setV7` (encode bytes 0–7), `stamp` (version + variant bits); every generator goes through them
- `bench/` — separate Go module with comparison benchmarks against google/uuid and gofrs/uuid

## Design Principles

- **No global configuration.** No SetRand, no swappable clock or random source. V4/V5/V8 are stateless. V7 uses a Generator with per-instance lock; the only package-level state is the default Generator behind NewV7/NewV7Batch.
- **No NullUUID.** Use `*UUID` pointer for SQL NULL.
- **Strict parsing by default.** `Parse()` = 36-char hyphenated only. `ParseLenient()` for other forms. `Scan` treats a `string` as text always; only a 16-byte `[]byte` is read as raw bytes (BINARY(16)).
- **Always crypto/rand.** No SetRand. Pool and Batch amortize cost without changing the CSPRNG source.
- **Zero-alloc hot paths.** NewV4, NewV5 (any name length), NewV7, Pool.NewV4, Pool.NewV7, Parse, ParseLenient, UnmarshalText, AppendText are all zero-alloc (enforced by the `*ZeroAlloc` tests). MarshalText and MarshalBinary must return a fresh slice, so they allocate once whenever the result is used; benchmarks keep the result in a sink so the compiler cannot stack-allocate it. FillV4, FillV7, and NewV5Bytes are zero-alloc too. NewV4Batch and NewV7Batch allocate only their result (make + Fill): FillV4 lets crypto/rand fill dst directly via unsafe.Slice; FillV7 reads the 8-byte rand_b values in one crypto/rand call into the upper half of dst and encodes front to back (see the proof comment in generate.go). Batch benchmarks keep their result in a sink, since an inlined make whose result is discarded would be stack-allocated.
- **Lookup table parsing.** 256-byte hex lookup table + pre-computed offset array; UnmarshalText parses []byte directly.
- **V7 uses RFC 9562 Method 3.** Sub-millisecond precision in rand_a via `frac * 4096 / 1_000_000`; monotonic counter fallback. Only reads 8 random bytes (rand_b) since bytes 0–7 are deterministic timestamp+sequence.
- **Pool amortizes crypto/rand.** Pool pre-generates 256 UUIDs (V4) or 256×8 random bytes (V7 rand_b) per refill. V4 pool: ~14x faster. V7 pool: ~2x faster (time.Now dominates). Batch APIs (NewV4Batch, NewV7Batch) amortize similarly for bulk generation (~30x for V4, ~15x for V7 at n=100).

## Go 1.24–1.27 Features Used

- `encoding.TextAppender` / `BinaryAppender` (1.24) — format.go
- `testing/synctest` (1.25) — generate_test.go fake clock for V7
- `crypto/rand` infallible (1.24) — NewV4 returns UUID, no error
- `testing/cryptotest.SetGlobalRandom` (1.26) — deterministic test randomness
- `errors.AsType[E]()` (1.26) — typed error matching in tests
- `encoding/json/v2` (1.27) — round-trip test in format_test.go
- `synctest.Sleep` (1.27) — generate_test.go

## Supported UUID Versions

V4 (random), V5 (SHA-1 name-based), V7 (timestamp+random), V8 (custom). No V1/V2/V3/V6 (legacy).

## Test Conventions

- Tests are internal (package `uuid`) except `example_test.go` (package `uuid_test`)
- Use `cryptotest.SetGlobalRandom(t, seed)` for deterministic randomness
- Use `synctest.Test(t, func(t *testing.T) { ... })` for fake-clock V7 tests
- V7 behaviour tests run against every source via `v7Sources()` (Generator, zero-value Generator, NewV7Batch, Generator.FillV7, Pool, zero-value Pool); add new V7 paths there
- Fuzz tests must round-trip: parse then re-parse the String() output
- Decoders never write through the receiver on error: decode into a local, assign on success
