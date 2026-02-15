package uuid

import (
	"cmp"
	"time"
)

// UUID is a 128-bit universally unique identifier per RFC 9562.
// It is a value type: comparable, copyable, and safe for use as a map key.
type UUID [16]byte

// Version represents the UUID version field.
type Version uint8

const (
	VersionNil Version = 0
	Version3   Version = 3
	Version4   Version = 4
	Version5   Version = 5
	Version7   Version = 7
	Version8   Version = 8
	VersionMax Version = 15
)

// String returns the version name.
func (v Version) String() string {
	switch v {
	case VersionNil:
		return "NIL"
	case Version3:
		return "V3"
	case Version4:
		return "V4"
	case Version5:
		return "V5"
	case Version7:
		return "V7"
	case Version8:
		return "V8"
	case VersionMax:
		return "MAX"
	default:
		return "unknown"
	}
}

// Variant represents the UUID variant field.
type Variant uint8

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

const hexDigits = "0123456789abcdef"

// String returns the standard 36-character hyphenated UUID representation:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (u UUID) String() string {
	var buf [36]byte
	encodeHex(buf[:], u)
	return string(buf[:])
}

// URN returns the UUID in URN form: urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (u UUID) URN() string {
	var buf [45]byte
	copy(buf[:9], "urn:uuid:")
	encodeHex(buf[9:], u)
	return string(buf[:])
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
// For non-V7 UUIDs, the returned time is meaningless.
func (u UUID) Time() time.Time {
	ms := int64(u[0])<<40 | int64(u[1])<<32 | int64(u[2])<<24 |
		int64(u[3])<<16 | int64(u[4])<<8 | int64(u[5])
	return time.UnixMilli(ms)
}

// Compare returns an integer comparing two UUIDs lexicographically.
// The result is 0 if a == b, -1 if a < b, and +1 if a > b.
// This is suitable for use with [slices.SortFunc].
func Compare(a, b UUID) int {
	return cmp.Compare(string(a[:]), string(b[:]))
}

// encodeHex writes the 36-byte hyphenated hex representation of u into dst.
// dst must be at least 36 bytes.
func encodeHex(dst []byte, u UUID) {
	hex := hexDigits
	dst[8] = '-'
	dst[13] = '-'
	dst[18] = '-'
	dst[23] = '-'
	dst[0] = hex[u[0]>>4]
	dst[1] = hex[u[0]&0x0f]
	dst[2] = hex[u[1]>>4]
	dst[3] = hex[u[1]&0x0f]
	dst[4] = hex[u[2]>>4]
	dst[5] = hex[u[2]&0x0f]
	dst[6] = hex[u[3]>>4]
	dst[7] = hex[u[3]&0x0f]
	dst[9] = hex[u[4]>>4]
	dst[10] = hex[u[4]&0x0f]
	dst[11] = hex[u[5]>>4]
	dst[12] = hex[u[5]&0x0f]
	dst[14] = hex[u[6]>>4]
	dst[15] = hex[u[6]&0x0f]
	dst[16] = hex[u[7]>>4]
	dst[17] = hex[u[7]&0x0f]
	dst[19] = hex[u[8]>>4]
	dst[20] = hex[u[8]&0x0f]
	dst[21] = hex[u[9]>>4]
	dst[22] = hex[u[9]&0x0f]
	dst[24] = hex[u[10]>>4]
	dst[25] = hex[u[10]&0x0f]
	dst[26] = hex[u[11]>>4]
	dst[27] = hex[u[11]&0x0f]
	dst[28] = hex[u[12]>>4]
	dst[29] = hex[u[12]&0x0f]
	dst[30] = hex[u[13]>>4]
	dst[31] = hex[u[13]&0x0f]
	dst[32] = hex[u[14]>>4]
	dst[33] = hex[u[14]&0x0f]
	dst[34] = hex[u[15]>>4]
	dst[35] = hex[u[15]&0x0f]
}
