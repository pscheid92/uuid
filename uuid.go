// Package uuid is a companion to the standard library uuid package (Go 1.27):
// it implements the parts of RFC 9562 and the conveniences that package
// leaves out.
//
// The standard library generates V4 and V7 UUIDs and parses and prints them.
// When that is all a program needs, use it. This package adds:
//   - V5 (SHA-1 name-based) and V8 (custom) UUIDs: [NewV5], [NewV8]
//   - accessors: [UUID.Version], [UUID.Variant], [UUID.Time]
//   - strict parsing with typed errors: [Parse], [ParseError], [ErrInvalid]
//   - V7 UUIDs for a given time, to backfill existing records: [NewV7At]
//   - V7 ordering per [Generator], kept even when the clock steps back
//   - bulk generation: [NewV4Batch], [NewV7Batch], [FillV4], [FillV7], [Pool]
//   - binary encoding: [UUID.MarshalBinary], [UUID.AppendBinary], [FromBytes]
//
// It also has [NewV4], [NewV7], [Parse], and [UUID.String], so code that
// needs the additions can use this package alone. Neither speed nor
// database/sql is a reason to choose it: single UUIDs cost the same in both
// packages, and database/sql handles the standard library's type itself.
//
// UUID is a 16-byte value type that is comparable and safe for use as a map key.
// The zero value is the Nil UUID (all zeros).
//
// # The standard library uuid package
//
// Both packages define UUID as [16]byte, so values convert in either
// direction at no cost. Convert where a dependency hands over or expects a
// standard library UUID:
//
//	import (
//		stdlib "uuid"
//
//		"github.com/pscheid92/uuid"
//	)
//
//	id := uuid.UUID(stdlib.New())    // standard library -> this package
//	std := stdlib.UUID(uuid.NewV7()) // this package -> standard library
//
// V7 UUIDs from the two packages are not strictly ordered relative to each
// other, so keys that must sort should come from one of them.
//
// The packages share the name uuid and many function names. goimports
// resolves a missing uuid import to the standard library unless other files
// in the package already import this one and use every uuid name the new
// file needs, so a new package or a first use of a function gets the
// standard library. Such a file may still compile, but its Parse is lenient
// and its UUIDs have no [UUID.Version] or [UUID.Time]. Check the import
// path, and alias the standard library package, as above, in files that use
// both.
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
// # SQL NULL handling
//
// UUID implements [database/sql.Scanner] and [database/sql/driver.Valuer].
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
//
// V1, V2, V3, and V6 are the legacy versions. This package does not generate
// them, but [Parse] accepts UUIDs of any version, so the constants are here
// to compare against [UUID.Version].
const (
	VNil Version = 0
	V1   Version = 1 // Gregorian time and node ID (legacy)
	V2   Version = 2 // DCE Security (legacy)
	V3   Version = 3 // MD5 name-based (legacy)
	V4   Version = 4
	V5   Version = 5
	V6   Version = 6 // reordered Gregorian time (legacy)
	V7   Version = 7
	V8   Version = 8
	VMax Version = 15
)

// String returns the version name, including the legacy versions this
// package does not generate.
func (v Version) String() string {
	switch v {
	case VNil:
		return "NIL"
	case V1:
		return "V1"
	case V2:
		return "V2"
	case V3:
		return "V3"
	case V4:
		return "V4"
	case V5:
		return "V5"
	case V6:
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
	// u is this call's own copy, so slicing it yields a fresh slice the
	// caller owns; this is faster than make and copy.
	return u[:]
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
