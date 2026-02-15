# UUID Library

Modern Go UUID library (RFC 9562) targeting Go 1.26. Module: `github.com/pscheid92/uuid`

## Commands

```bash
go test ./...                           # run all tests
go test -race ./...                     # run with race detector
go vet ./...                            # static analysis
go test -bench=. -benchmem ./...        # benchmarks with alloc stats
go test -fuzz='^FuzzParse$' -fuzztime=30s ./...       # fuzz Parse
go test -fuzz=FuzzParseLenient -fuzztime=30s ./...    # fuzz ParseLenient
cd bench && go test -bench=. -benchmem ./...          # comparison benchmarks vs google/uuid, gofrs/uuid
```

## Architecture

Single flat package at the module root. Each file has a focused responsibility:

- `uuid.go` — package doc, UUID type, Nil/Max, Namespace constants, Version/Variant types (VNil/V3/V4/V5/V7/V8/VMax), accessors (Version/Variant/IsNil/Bytes/Time/Compare)
- `parse.go` — Parse (strict 36-char), ParseLenient (URN/braced/compact), MustParse, FromBytes; hex lookup table + offset array; ParseError, LengthError
- `format.go` — String, URN, encodeHex, AppendText/Binary, Marshal/Unmarshal (Text + Binary); Scan (database/sql.Scanner), Value (driver.Valuer)
- `generate.go` — NewV3/V4/V5/V7/V8, NewV4Batch, Generator type with per-instance V7 monotonicity (RFC 9562 Method 3) and NewV7Batch, Pool type with buffered NewV4/NewV7, hash.Cloner setup
- `bench/` — separate Go module with comparison benchmarks against google/uuid and gofrs/uuid

## Design Principles

- **No global mutable state.** V3/V4/V5/V8 are pure functions. V7 uses a Generator with per-instance lock.
- **No NullUUID.** Use `*UUID` pointer for SQL NULL.
- **Strict parsing by default.** `Parse()` = 36-char hyphenated only. `ParseLenient()` for other forms.
- **Always crypto/rand.** No SetRand. Pool and Batch amortize cost without changing the CSPRNG source.
- **Zero-alloc hot paths.** NewV4, NewV7, Pool.NewV4, Pool.NewV7, Parse, UnmarshalText, AppendText, MarshalText are all zero-alloc.
- **Lookup table parsing.** 256-byte hex lookup table + pre-computed offset array; UnmarshalText parses []byte directly.
- **hash.Cloner optimization.** V3/V5 pre-initialize hash states for standard namespaces at init, clone per call.
- **V7 uses RFC 9562 Method 3.** Sub-millisecond precision in rand_a via `frac * 4096 / 1_000_000`; monotonic counter fallback. Only reads 8 random bytes (rand_b) since bytes 0–7 are deterministic timestamp+sequence.
- **Pool amortizes crypto/rand.** Pool pre-generates 256 UUIDs (V4) or 256×8 random bytes (V7 rand_b) per refill. V4 pool: ~14x faster. V7 pool: ~2x faster (time.Now dominates). Batch APIs (NewV4Batch, NewV7Batch) amortize similarly for bulk generation (~25x for V4, ~13x for V7 at n=100).

## Go 1.24–1.26 Features Used

- `encoding.TextAppender` / `BinaryAppender` (1.24) — format.go
- `hash.Cloner` (1.25) — generate.go namespace hash cloning
- `testing/synctest` (1.25) — generate_test.go fake clock for V7
- `crypto/rand` infallible (1.26) — NewV4 returns UUID, no error
- `testing/cryptotest.SetGlobalRandom` (1.26) — deterministic test randomness
- `errors.AsType[E]()` (1.26) — typed error matching in tests

## Supported UUID Versions

V3 (MD5), V4 (random), V5 (SHA-1), V7 (timestamp+random), V8 (custom). No V1/V2/V6 (legacy).

## Test Conventions

- Tests are internal (package `uuid`) except `example_test.go` (package `uuid_test`)
- Use `cryptotest.SetGlobalRandom(t, seed)` for deterministic randomness
- Use `synctest.Test(t, func(t *testing.T) { ... })` for fake-clock V7 tests
- Fuzz tests must round-trip: parse then re-parse the String() output
