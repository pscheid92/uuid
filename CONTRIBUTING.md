# Contributing

Thank you for considering contributing to this project!

## Requirements

- **Go 1.26+**: this library uses features from Go 1.24 through 1.26
- **Zero external dependencies**: only the Go standard library is allowed
- **100% test coverage**: verify with:
  ```bash
  go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
  ```

## Development Commands

```bash
go test ./...                       # run all tests
go test -race ./...                 # run with race detector
go vet ./...                        # static analysis
go test -bench=. -benchmem ./...    # benchmarks
go test -fuzz='^FuzzParse$' -fuzztime=30s ./...     # fuzz Parse
go test -fuzz=FuzzParseLenient -fuzztime=30s ./...  # fuzz ParseLenient
```

## Test Conventions

- Tests are internal (package `uuid`) except `example_test.go` (package `uuid_test`)
- Use `cryptotest.SetGlobalRandom(t, seed)` for deterministic randomness
- Use `synctest.Test(t, func(t *testing.T) { ... })` for fake-clock V7 tests
- Fuzz tests must round-trip: parse then re-parse the `String()` output

## Pull Request Checklist

- [ ] `go test -race ./...` passes
- [ ] `go vet ./...` is clean
- [ ] 100% test coverage maintained
- [ ] No new external dependencies added
- [ ] Documentation updated if public API changed

## Design Principles

See [Why This Library?](README.md#why-this-library) in the README for the design principles guiding this project.
