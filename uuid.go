// Package uuid implements UUID generation and parsing per RFC 9562.
//
// Supported versions:
//   - V4 (Random): most common
//   - V5 (SHA-1 name-based): deterministic, canonical IDs
//   - V7 (Unix timestamp + random): recommended for new systems
//   - V8 (Custom/experimental): user-provided data with version+variant bits
//
// UUID is a 16-byte value type that is comparable and safe for use as a map key.
// The zero value is the Nil UUID (all zeros).
//
// # Generation
//
// Stateless functions require no configuration:
//
//	id := uuid.NewV4()                              // random
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
// The rule follows the source of the text. [Parse], [MustParse], and
// [UUID.UnmarshalText] (and therefore JSON, XML, and other text encodings)
// accept only the 36-character form. [ParseLenient] and [UUID.Scan]
// (database/sql) accept all four. The standard library and google/uuid parse
// leniently everywhere; code moving from them should call [ParseLenient]
// where it relied on that. The LenientJSON example shows how to accept every
// form in JSON.
//
// # The standard library uuid package
//
// Go 1.27 added a standard library package that is also named uuid, with the
// same function names. goimports resolves a missing uuid import to the
// standard library unless other files in the package already import this
// one and use every uuid name the new file needs, so a new package or a
// first use of a function gets the standard library. Such a file may still
// compile, but its UUIDs have no [UUID.Scan] or [UUID.Value]. Check the
// import path, and alias the standard library package in files that use
// both:
//
//	import (
//		stdlib "uuid"
//
//		"github.com/pscheid92/uuid"
//	)
//
// Both types are [16]byte, so uuid.UUID(s) and stdlib.UUID(u) convert freely.
//
// # SQL NULL handling
//
// Instead of a separate NullUUID type, use a *UUID pointer:
//
//	var id *uuid.UUID  // nil = SQL NULL
package uuid

import (
	"bytes"
	"time"
)

// UUID is a 128-bit universally unique identifier per RFC 9562.
// It is a value type: comparable, copyable, and safe for use as a map key.
type UUID [16]byte

// Nil is the zero-value UUID (all zeros).
//
// Nil, Max, and the Namespace values are variables only because Go has no
// array constants. Treat them as read-only: writing to one changes it for
// every user in the process, including [NewV5] callers.
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

// Version represents the UUID version field.
type Version uint8

// UUID version constants.
const (
	VNil Version = 0
	V4   Version = 4
	V5   Version = 5
	V7   Version = 7
	V8   Version = 8
	VMax Version = 15
)

// String returns the version name. Legacy versions (V1, V2, V3, V6) are
// named even though this package does not generate them, since [Parse]
// accepts UUIDs of any version.
func (v Version) String() string {
	switch v {
	case VNil:
		return "NIL"
	case 1:
		return "V1"
	case 2:
		return "V2"
	case 3:
		return "V3"
	case V4:
		return "V4"
	case V5:
		return "V5"
	case 6:
		return "V6"
	case V7:
		return "V7"
	case V8:
		return "V8"
	case VMax:
		return "MAX"
	default:
		return "unknown"
	}
}

// Variant represents the UUID variant field.
type Variant uint8

// UUID variant constants.
const (
	VariantNCS       Variant = 0 // NCS backward compatibility
	VariantRFC9562   Variant = 1 // RFC 9562 (formerly RFC 4122)
	VariantMicrosoft Variant = 2 // Microsoft backward compatibility
	VariantFuture    Variant = 3 // Reserved for future definition
)

// String returns the variant name.
func (v Variant) String() string {
	switch v {
	case VariantNCS:
		return "NCS"
	case VariantRFC9562:
		return "RFC9562"
	case VariantMicrosoft:
		return "Microsoft"
	case VariantFuture:
		return "Future"
	default:
		return "unknown"
	}
}

// Version returns the UUID version (bits 48–51).
func (u UUID) Version() Version {
	return Version(u[6] >> 4)
}

// Variant returns the UUID variant (bits 64–65).
func (u UUID) Variant() Variant {
	b := u[8]
	switch {
	case b&0x80 == 0x00:
		return VariantNCS
	case b&0xc0 == 0x80:
		return VariantRFC9562
	case b&0xe0 == 0xc0:
		return VariantMicrosoft
	default:
		return VariantFuture
	}
}

// IsNil reports whether u is the zero-value (Nil) UUID.
func (u UUID) IsNil() bool {
	return u == Nil
}

// Bytes returns a copy of the UUID as a 16-byte slice.
func (u UUID) Bytes() []byte {
	b := make([]byte, 16)
	copy(b, u[:])
	return b
}

// Time extracts the millisecond-precision Unix timestamp from a V7 UUID.
// The boolean is false, and the time is the zero value, if u is not an
// RFC 9562 Version 7 UUID; other versions either carry no timestamp or lay
// it out differently, so decoding them would return a plausible but wrong
// time. The variant is checked too, because the version field is only
// defined for [VariantRFC9562]: a Microsoft GUID whose version bits happen
// to read 7 carries no timestamp.
//
// For V7 UUIDs generated by this package under sustained bursts, the
// timestamp may be marginally ahead of the actual creation time due to
// the monotonic counter (see [Generator.NewV7]).
func (u UUID) Time() (time.Time, bool) {
	if u.Version() != V7 || u.Variant() != VariantRFC9562 {
		return time.Time{}, false
	}
	ms := int64(u[0])<<40 | int64(u[1])<<32 | int64(u[2])<<24 |
		int64(u[3])<<16 | int64(u[4])<<8 | int64(u[5])
	return time.UnixMilli(ms), true
}

// Compare returns an integer comparing u and v lexicographically.
// The result is 0 if u == v, -1 if u < v, and +1 if u > v.
// It matches the Compare method of the standard library uuid package.
func (u UUID) Compare(v UUID) int {
	return bytes.Compare(u[:], v[:])
}

// Compare returns an integer comparing two UUIDs lexicographically.
// The result is 0 if a == b, -1 if a < b, and +1 if a > b.
// It is equivalent to a.Compare(b) and suitable for use with [slices.SortFunc].
func Compare(a, b UUID) int {
	return a.Compare(b)
}
