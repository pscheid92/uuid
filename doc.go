// Package uuid implements UUID generation and parsing per RFC 9562.
//
// Supported versions:
//   - Version 3 (MD5 name-based): deterministic, canonical IDs
//   - Version 4 (Random): most common
//   - Version 5 (SHA-1 name-based): deterministic, preferred over V3
//   - Version 7 (Unix timestamp + random): recommended for new systems
//   - Version 8 (Custom/experimental): user-provided data with version+variant bits
//
// UUID is a 16-byte value type that is comparable and safe for use as a map key.
// The zero value is the Nil UUID (all zeros).
//
// # Generation
//
// Stateless functions require no configuration:
//
//	id := uuid.NewV4()                              // random
//	id := uuid.NewV3(uuid.NamespaceDNS, "example")  // deterministic (MD5)
//	id := uuid.NewV5(uuid.NamespaceDNS, "example")  // deterministic (SHA-1)
//
// For V7 UUIDs with per-instance monotonicity, use a [Generator]:
//
//	gen := uuid.NewGenerator()
//	id := gen.NewV7()
//
// A package-level convenience function is also available:
//
//	id := uuid.NewV7()
//
// # Parsing
//
// [Parse] is strict: it only accepts the standard 36-character hyphenated form.
// [ParseLenient] additionally accepts URN, braced, and compact (32-hex) forms.
//
//	id, err := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
//	id, err := uuid.ParseLenient("urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8")
//
// # SQL NULL handling
//
// Instead of a separate NullUUID type, use a *UUID pointer:
//
//	var id *uuid.UUID  // nil = SQL NULL
package uuid

// Nil is the zero-value UUID (all zeros).
var Nil UUID

// Max is the maximum UUID (all 0xFF bytes), defined in RFC 9562 Section 5.10.
var Max = UUID{
	0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
}

// RFC 9562 Appendix C pre-defined namespace UUIDs.
var (
	NamespaceDNS  = UUID{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	NamespaceURL  = UUID{0x6b, 0xa7, 0xb8, 0x11, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	NamespaceOID  = UUID{0x6b, 0xa7, 0xb8, 0x12, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	NamespaceX500 = UUID{0x6b, 0xa7, 0xb8, 0x14, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
)
