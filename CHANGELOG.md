# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-02-15

### Added

- UUID type as `[16]byte` value type with `Nil` and `Max` constants
- UUID generation: `NewV3` (MD5), `NewV4` (random), `NewV5` (SHA-1), `NewV7` (timestamp+random), `NewV8` (custom)
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
- `hash.Cloner` optimization for V3/V5 namespace hash states
- Zero-alloc hot paths for NewV4, NewV7, Parse, MarshalText, UnmarshalText
- 100% test coverage including fuzz tests

[0.1.0]: https://github.com/pscheid92/uuid/releases/tag/v0.1.0
