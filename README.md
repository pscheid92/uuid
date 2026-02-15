# uuid

A modern Go UUID library implementing [RFC 9562](https://www.rfc-editor.org/rfc/rfc9562). Zero external dependencies. Requires Go 1.26+.

```
go get github.com/pscheid92/uuid
```

## Supported Versions

| Version | Description | Function |
|---------|-------------|----------|
| V3 | Deterministic (MD5) | `NewV3(namespace, name)` |
| V4 | Random | `NewV4()` |
| V5 | Deterministic (SHA-1) | `NewV5(namespace, name)` |
| V7 | Timestamp + random | `NewV7()` |
| V8 | Custom data | `NewV8(data)` |

## Usage

### Generate UUIDs

```go
import "github.com/pscheid92/uuid"

// Random (V4) — most common
id := uuid.NewV4()

// Timestamp-ordered (V7) — recommended for new systems, database-friendly
id := uuid.NewV7()

// Deterministic (V5, SHA-1) — same inputs always produce the same UUID
id := uuid.NewV5(uuid.NamespaceDNS, "www.example.com")
// 2ed6657d-e927-568b-95e1-2665a8aea6a2

// Deterministic (V3, MD5) — prefer V5 over V3
id := uuid.NewV3(uuid.NamespaceDNS, "www.example.com")
// 5df41881-3aed-3515-88a7-2f4a814cf09e
```

### V7 with Monotonicity Guarantees

Multiple V7 UUIDs generated within the same millisecond are guaranteed to sort in creation order within a single `Generator`:

```go
gen := uuid.NewGenerator()
id1 := gen.NewV7()
id2 := gen.NewV7() // guaranteed id1 < id2 even within the same millisecond
```

The package-level `uuid.NewV7()` uses a default shared generator.

### Parse

`Parse` is strict — it only accepts the standard 36-character hyphenated form:

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

### String Representations

```go
id := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

id.String() // "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
id.URN()    // "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"
```

### Inspect Properties

```go
id := uuid.NewV7()

id.Version() // uuid.Version7
id.Variant()  // uuid.VariantRFC9562
id.IsNil()    // false
id.Time()     // time.Time (millisecond precision, V7 only)
```

### JSON

UUID implements `encoding.TextMarshaler` and `encoding.TextUnmarshaler`, so JSON works automatically:

```go
type User struct {
    ID   uuid.UUID  `json:"id"`
    Name string     `json:"name"`
}

// Marshals to:   {"id":"550e8400-e29b-41d4-a716-446655440000","name":"alice"}
// Unmarshals from the same format
```

Use a pointer for nullable fields:

```go
type Event struct {
    ID       uuid.UUID  `json:"id"`
    ParentID *uuid.UUID `json:"parent_id"` // null in JSON when nil
}
```

### SQL / Database

UUID implements `database/sql.Scanner` and `driver.Valuer`. Use a pointer for nullable columns:

```go
var id uuid.UUID
err := row.Scan(&id)

// Nullable column
var id *uuid.UUID // nil = SQL NULL
err := row.Scan(&id)
```

### Sorting

```go
ids := []uuid.UUID{id3, id1, id2}
slices.SortFunc(ids, uuid.Compare)
```

### Predefined Namespace UUIDs

For use with `NewV3` and `NewV5` (RFC 9562 Appendix C):

```go
uuid.NamespaceDNS   // 6ba7b810-9dad-11d1-80b4-00c04fd430c8
uuid.NamespaceURL   // 6ba7b811-9dad-11d1-80b4-00c04fd430c8
uuid.NamespaceOID   // 6ba7b812-9dad-11d1-80b4-00c04fd430c8
uuid.NamespaceX500  // 6ba7b814-9dad-11d1-80b4-00c04fd430c8
```

## Design

- **Value type** — `UUID` is `[16]byte`, comparable, copyable, safe as map key.
- **No global mutable state** — V3/V4/V5/V8 are pure functions. V7 uses per-instance `Generator`.
- **Strict parsing by default** — `Parse` rejects anything that isn't `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`.
- **No NullUUID** — Use `*UUID` for nullable values in SQL and JSON.
- **Zero-alloc hot paths** — NewV4, NewV7, Parse, MarshalText, and UnmarshalText allocate nothing.
- **Zero dependencies** — Only uses the Go standard library.
