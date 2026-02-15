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

- `doc.go` — package comment only
- `uuid.go` — UUID type, Nil/Max, Namespace constants, Version/Variant types (VNil/V3/V4/V5/V7/V8/VMax), accessors (Version/Variant/IsNil/Bytes/Time/Compare)
- `errors.go` — ParseError, LengthError (use `errors.AsType[*ParseError](err)`)
- `parse.go` — Parse (strict 36-char), ParseLenient (URN/braced/compact), MustParse, FromBytes; hex lookup table + offset array
- `format.go` — String, URN, encodeHex, AppendText/Binary, Marshal/Unmarshal (Text + Binary)
- `generate.go` — NewV3/V4/V5/V7/V8, Generator type with per-instance V7 monotonicity (RFC 9562 Method 3), hash.Cloner setup
- `sql.go` — Scan (database/sql.Scanner), Value (driver.Valuer)
- `bench/` — separate Go module with comparison benchmarks against google/uuid and gofrs/uuid

## Design Principles

- **No global mutable state.** V3/V4/V5/V8 are pure functions. V7 uses a Generator with per-instance lock.
- **No NullUUID.** Use `*UUID` pointer for SQL NULL.
- **Strict parsing by default.** `Parse()` = 36-char hyphenated only. `ParseLenient()` for other forms.
- **No SetRand/EnableRandPool.** Go 1.26 crypto/rand always uses system CSPRNG.
- **Zero-alloc hot paths.** NewV4, NewV7, Parse, UnmarshalText, AppendText, MarshalText are all zero-alloc.
- **Lookup table parsing.** 256-byte hex lookup table + pre-computed offset array; UnmarshalText parses []byte directly.
- **hash.Cloner optimization.** V3/V5 pre-initialize hash states for standard namespaces at init, clone per call.
- **V7 uses RFC 9562 Method 3.** Sub-millisecond precision in rand_a via `frac * 4096 / 1_000_000`; monotonic counter fallback.

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
