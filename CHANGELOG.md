# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `NewV7At(t time.Time)` builds a V7 UUID whose timestamp fields encode `t`, for backfilling rows that already have a creation time. Pure function: it never touches `Generator` or `Pool` monotonic state. Panics for times outside the 48-bit range, including the zero `time.Time`.
- `UUID.Value` example and docs section showing a `BinaryUUID` wrapper for BINARY(16) columns (MySQL, MariaDB)

### Changed

- Zero-alloc tests now cover `NewV4`, `NewV7`, `NewV7At`, `Generator.NewV7`, `Pool.NewV4`, and `Pool.NewV7`; they skip under the race detector, where `crypto/rand.Read` allocates on Linux, and CI runs them separately without it
- **Breaking:** `UUID.Time()` now returns `(time.Time, bool)`. The boolean is false, and the time is the zero value, for any non-V7 UUID. Previously a V1 or V6 UUID decoded to a plausible but wrong time, and V4 to garbage. Migrate with `t, _ := id.Time()` where the version is already known.

- `Generator.NewV7Batch` now reserves its sequence range under the lock and encodes the UUIDs outside it, so a large batch no longer blocks concurrent `NewV7` callers on the same generator (~2x faster single-UUID generation while a 10k batch runs in parallel)
- `Compare` uses `bytes.Compare` instead of string conversions (~3x faster)
- `ParseLenient` hands the hex decoder a 36-byte window instead of an offset, letting the compiler elide bounds checks (~20% faster); `Parse` is unchanged
- `AppendText` uses `slices.Grow` instead of a private helper
- `ParseError.Input` truncation no longer splits a multi-byte UTF-8 rune
- URN and braced `ParseLenient` errors are checked to report the full input
- Documented that each `Pool` keeps V7 monotonic state independent of the package-level generator and other pools

## [0.4.0] - 2026-09-10

### Security

- `Pool` zero value is now safe to use. Previously `var p uuid.Pool` handed out Nil UUIDs (V4) and UUIDs with an all-zero, predictable `rand_b` field (V7) for the first 256 calls; only `NewPool()` armed the refill. If you construct pools without `NewPool`, upgrade.

### Changed

- `NewV5` is zero-alloc for names up to 240 bytes (was 4 allocs) and ~40% faster; `hash.Cloner` pre-hashing removed
- `NewV4Batch` fills the result directly from `crypto/rand` (1 alloc instead of 2)
- `Version.String` names legacy versions V1, V2, V3, and V6 instead of returning "unknown"
- CI reads the Go version from `go.mod` instead of hard-coding it
- Added `encoding/json/v2` round-trip test

## [0.3.0] - 2026-09-10

### Added

- Package-level `NewV7Batch(n)` using the default generator, mirroring `NewV4Batch`
- `Scan` accepts a 16-byte raw `string` (e.g. a `BINARY(16)` column delivered as string)

### Changed

- `NewV4Batch` and `NewV7Batch` return nil for `n <= 0` instead of panicking on negative counts
- Minimum Go version raised to 1.27; bench module dependencies updated (gofrs/uuid v5.5.1)
- `Scan(nil)` returns a dedicated error suggesting `*UUID` for nullable columns

### Fixed

- `ParseLenient` accepts the `urn:uuid:` prefix case-insensitively per RFC 8141

### Security

- `ParseError.Input` is truncated to 64 bytes so unbounded inputs are not copied into error messages and logs
- Documented that `Pool` is not fork-safe or VM-clone-safe (godoc, SECURITY.md, docs/advanced.md)
- Documented that V7 timestamps may run slightly ahead of the wall clock under sustained bursts
- CI runs `govulncheck` and vets the bench module

## [0.2.0] - 2026-03-14

### Removed

- `NewV3` and `V3` version constant — V3 (MD5) is superseded by V5 (SHA-1) per RFC 9562

### Changed

- `hash.Cloner` optimization now only covers SHA-1 (V5), MD5 removed
- `.golangci.yml` gosec exclusions narrowed to G505 only (SHA-1)

## [0.1.0] - 2026-02-15

### Added

- UUID type as `[16]byte` value type with `Nil` and `Max` constants
- UUID generation: `NewV4` (random), `NewV5` (SHA-1), `NewV7` (timestamp+random), `NewV8` (custom)
- `Generator` type with per-instance V7 monotonicity (RFC 9562 Method 3)
- Predefined namespace UUIDs: `NamespaceDNS`, `NamespaceURL`, `NamespaceOID`, `NamespaceX500`
- `Parse` (strict 36-char hyphenated) and `ParseLenient` (URN, braced, compact forms)
- `MustParse` for package-level constants
- `FromBytes` constructor from byte slices
- `String` and `URN` formatting methods
- `Version`, `Variant`, `IsNil`, `Bytes`, `Time`, `Compare` accessors
- `encoding.TextMarshaler` / `TextUnmarshaler` for JSON support
- `encoding.BinaryMarshaler` / `BinaryUnmarshaler` for binary protocols
- `encoding.TextAppender` / `BinaryAppender` (Go 1.24) for zero-alloc formatting
- `database/sql.Scanner` and `driver.Valuer` for SQL support
- `hash.Cloner` optimization for V5 namespace hash states
- `Pool` type with `NewV4` and `NewV7` for amortized `crypto/rand` overhead
- `NewV4Batch(n)` and `Generator.NewV7Batch(n)` for bulk UUID generation
- Zero-alloc hot paths for NewV4, NewV7, Pool.NewV4, Pool.NewV7, Parse, MarshalText, UnmarshalText
- 100% test coverage including fuzz tests

[Unreleased]: https://github.com/pscheid92/uuid/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/pscheid92/uuid/releases/tag/v0.4.0
[0.3.0]: https://github.com/pscheid92/uuid/releases/tag/v0.3.0
[0.2.0]: https://github.com/pscheid92/uuid/releases/tag/v0.2.0
[0.1.0]: https://github.com/pscheid92/uuid/releases/tag/v0.1.0
