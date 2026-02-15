package uuid

import "fmt"

// xvalues maps hex character bytes to their values; 0xff marks invalid.
var xvalues = [256]byte{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
}

// xtob converts two hex characters into a byte.
func xtob(x1, x2 byte) (byte, bool) {
	b1 := xvalues[x1]
	b2 := xvalues[x2]
	return b1<<4 | b2, b1 != 0xff && b2 != 0xff
}

// hexOffsets maps each UUID byte index to the position of its high hex digit
// within the 36-char hyphenated format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx).
var hexOffsets = [16]int{
	0, 2, 4, 6, // bytes 0–3
	9, 11, // bytes 4–5
	14, 16, // bytes 6–7
	19, 21, // bytes 8–9
	24, 26, 28, 30, 32, 34, // bytes 10–15
}

// Parse parses a UUID from the standard 36-character hyphenated form:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
//
// For URN, braced, or compact (32-hex) forms, use [ParseLenient].
func Parse(s string) (UUID, error) {
	if len(s) != 36 {
		return Nil, &ParseError{Input: s, Msg: "expected 36-character hyphenated format"}
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return Nil, &ParseError{Input: s, Msg: "expected hyphens at positions 8, 13, 18, 23"}
	}
	var u UUID
	for i, x := range hexOffsets {
		v, ok := xtob(s[x], s[x+1])
		if !ok {
			return Nil, &ParseError{Input: s, Msg: "invalid hex character"}
		}
		u[i] = v
	}
	return u, nil
}

// ParseLenient parses a UUID from any of these forms:
//   - Standard:  xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 chars)
//   - URN:       urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (45 chars)
//   - Braced:    {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx} (38 chars)
//   - Compact:   xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx (32 chars)
func ParseLenient(s string) (UUID, error) {
	switch len(s) {
	case 36: // standard
		if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
			return Nil, &ParseError{Input: s, Msg: "expected hyphens at positions 8, 13, 18, 23"}
		}
		return parseHex(s, 0)

	case 45: // urn:uuid:
		if s[:9] != "urn:uuid:" {
			return Nil, &ParseError{Input: s, Msg: "expected urn:uuid: prefix"}
		}
		if s[17] != '-' || s[22] != '-' || s[27] != '-' || s[32] != '-' {
			return Nil, &ParseError{Input: s, Msg: "expected hyphens in UUID portion"}
		}
		return parseHex(s, 9)

	case 38: // {braced}
		if s[0] != '{' || s[37] != '}' {
			return Nil, &ParseError{Input: s, Msg: "expected braces"}
		}
		if s[9] != '-' || s[14] != '-' || s[19] != '-' || s[24] != '-' {
			return Nil, &ParseError{Input: s, Msg: "expected hyphens in UUID portion"}
		}
		return parseHex(s, 1)

	case 32: // compact (no hyphens)
		return parseCompact(s)

	default:
		return Nil, &ParseError{Input: s, Msg: "unrecognized UUID format"}
	}
}

// MustParse is like [Parse] but panics if the string cannot be parsed.
// It simplifies initialization of global variables holding UUIDs.
func MustParse(s string) UUID {
	id, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

// FromBytes creates a UUID from a 16-byte slice.
func FromBytes(b []byte) (UUID, error) {
	if len(b) != 16 {
		return Nil, &LengthError{Got: len(b), Want: "16 bytes"}
	}
	return UUID(b), nil
}

// parseHex decodes the 32 hex digits from s starting at offset,
// skipping the hyphens at the standard positions.
func parseHex(s string, offset int) (UUID, error) {
	var u UUID
	for i, x := range hexOffsets {
		x += offset
		v, ok := xtob(s[x], s[x+1])
		if !ok {
			return Nil, &ParseError{Input: s, Msg: "invalid hex character"}
		}
		u[i] = v
	}
	return u, nil
}

// parseCompact decodes a 32-character hex string with no hyphens.
func parseCompact(s string) (UUID, error) {
	var u UUID
	for i := range 16 {
		v, ok := xtob(s[i*2], s[i*2+1])
		if !ok {
			return Nil, &ParseError{Input: s, Msg: "invalid hex character"}
		}
		u[i] = v
	}
	return u, nil
}

// parseHexBytes decodes 32 hex digits from b starting at offset,
// writing the result into u. Used by UnmarshalText to avoid string conversion.
func parseHexBytes(u *UUID, b []byte, offset int) bool {
	for i, x := range hexOffsets {
		x += offset
		v, ok := xtob(b[x], b[x+1])
		if !ok {
			return false
		}
		u[i] = v
	}
	return true
}

// ParseError is returned when a UUID string cannot be parsed.
//
// Use [errors.AsType] to check for this error:
//
//	if perr, ok := errors.AsType[*ParseError](err); ok {
//	    fmt.Println(perr.Input)
//	}
type ParseError struct {
	Input string // the string that failed to parse
	Msg   string // description of the problem
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("uuid: parsing %q: %s", e.Input, e.Msg)
}

// LengthError is returned when the input has an unexpected byte length.
//
// Use [errors.AsType] to check for this error:
//
//	if lerr, ok := errors.AsType[*LengthError](err); ok {
//	    fmt.Println(lerr.Got, lerr.Want)
//	}
type LengthError struct {
	Got  int    // the actual length
	Want string // description of expected length
}

func (e *LengthError) Error() string {
	return fmt.Sprintf("uuid: unexpected length %d, want %s", e.Got, e.Want)
}
