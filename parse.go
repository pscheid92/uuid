package uuid

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

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
		return Nil, &ParseError{Input: errInput(s), Msg: "expected 36-character hyphenated format"}
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return Nil, hyphenError(s, 0)
	}
	var u UUID
	for i, x := range hexOffsets {
		v, ok := xtob(s[x], s[x+1])
		if !ok {
			return Nil, hexError(s, x)
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
		return parseHex(s, s, 0)

	case 45: // urn:uuid:
		// The urn scheme and uuid namespace are case-insensitive (RFC 8141).
		if !strings.EqualFold(s[:9], "urn:uuid:") {
			return Nil, &ParseError{Input: s, Msg: "expected urn:uuid: prefix"}
		}
		return parseHex(s[9:45], s, 9)

	case 38: // {braced}
		if s[0] != '{' || s[37] != '}' {
			return Nil, &ParseError{Input: s, Msg: "expected braces"}
		}
		return parseHex(s[1:37], s, 1)

	case 32: // compact (no hyphens)
		if u, ok := parseCompact(s); ok {
			return u, nil
		}
		return Nil, hexError(s, 0)

	default:
		return Nil, &ParseError{Input: errInput(s), Msg: "unrecognized UUID format"}
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

// parseHex decodes the 36-character hyphenated window s, which the caller
// has sliced out of the full input at offset off. Passing a window instead
// of indexing input at off keeps every index a constant, so the compiler can
// elide bounds checks; off is used only to report error positions relative
// to the full input.
func parseHex(s, input string, off int) (UUID, error) {
	_ = s[35]
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return Nil, hyphenError(input, off)
	}
	var u UUID
	for i, x := range hexOffsets {
		v, ok := xtob(s[x], s[x+1])
		if !ok {
			return Nil, hexError(input, off+x)
		}
		u[i] = v
	}
	return u, nil
}

// parseCompact decodes a 32-character hex string with no hyphens. It
// reports failure as a bool, leaving the error to the caller, so that it
// stays cheap enough for the compiler to inline into ParseLenient.
func parseCompact(s string) (UUID, bool) {
	var u UUID
	for i := range 16 {
		v, ok := xtob(s[i*2], s[i*2+1])
		if !ok {
			return Nil, false
		}
		u[i] = v
	}
	return u, true
}

// hyphenError reports the first missing hyphen in the 36-character
// hyphenated form that starts at input[off]. Callers invoke it only after
// one of the four hyphen checks has failed.
func hyphenError(input string, off int) *ParseError {
	var p int
	for _, h := range [...]int{8, 13, 18, 23} {
		p = off + h
		if input[p] != '-' {
			break
		}
	}
	return &ParseError{Input: input, Msg: fmt.Sprintf("expected '-' at position %d", p)}
}

// hexError reports the first invalid hex character in input at or after
// position p. Callers pass the start of a hex pair that xtob rejected, or
// the start of a compact input that failed to decode, so the scan stops
// within that pair or that input.
func hexError(input string, p int) *ParseError {
	for xvalues[input[p]] != 0xff {
		p++
	}
	return &ParseError{Input: input, Msg: fmt.Sprintf("invalid hex character at position %d", p)}
}

// maxErrInputLen bounds how much of a failed input is retained in a
// ParseError, so that arbitrarily large inputs are not copied into error
// messages and logs.
const maxErrInputLen = 64

// errInput returns s truncated to at most maxErrInputLen bytes for
// inclusion in a ParseError. The cut is moved back to a rune boundary so
// a multi-byte UTF-8 sequence is never split.
func errInput(s string) string {
	if len(s) <= maxErrInputLen {
		return s
	}
	n := maxErrInputLen
	for n > maxErrInputLen-utf8.UTFMax && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n] + "..."
}

// errInputBytes is errInput for []byte, bounding the slice before the
// string conversion to avoid copying large inputs.
func errInputBytes(b []byte) string {
	if len(b) <= maxErrInputLen {
		return string(b)
	}
	return errInput(string(b[:maxErrInputLen+1]))
}

// ErrInvalid matches, through [errors.Is], every error this package returns
// for input that is not a valid UUID: a [ParseError] from parsing, text
// decoding, or [UUID.Scan], and a [LengthError] from [FromBytes] or binary
// decoding. It does not match Scan's errors for SQL NULL or an unsupported
// source type, which are not malformed UUIDs.
//
//	if errors.Is(err, uuid.ErrInvalid) {
//	    http.Error(w, "invalid id", http.StatusBadRequest)
//	}
//
// Use [errors.AsType] on the concrete types for the offending input.
var ErrInvalid = errors.New("uuid: invalid UUID")

// ParseError is returned when a UUID string cannot be parsed.
// It matches [ErrInvalid].
//
// Use [errors.AsType] to check for this error:
//
//	if perr, ok := errors.AsType[*ParseError](err); ok {
//	    fmt.Println(perr.Input)
//	}
type ParseError struct {
	Input string // the string that failed to parse, truncated to 64 bytes if longer
	Msg   string // description of the problem; positions are byte offsets into Input
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("uuid: parsing %q: %s", e.Input, e.Msg)
}

// Is reports whether target is [ErrInvalid].
func (e *ParseError) Is(target error) bool { return target == ErrInvalid }

// LengthError is returned when the input has an unexpected byte length.
// It matches [ErrInvalid].
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

// Is reports whether target is [ErrInvalid].
func (e *LengthError) Is(target error) bool { return target == ErrInvalid }
