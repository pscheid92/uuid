package uuid

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"slices"
)

const hexDigits = "0123456789abcdef"

// String returns the standard 36-character hyphenated UUID representation:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (u UUID) String() string {
	var buf [36]byte
	encodeHex(buf[:], u)
	return string(buf[:])
}

// Format implements [fmt.Formatter] so that %x and %X print the 32 hex
// digits of u without hyphens, formatted like a [16]byte: "% x" spaces the
// bytes and "%#x" adds a 0x prefix. Without it, %x would hex-encode the
// 36-character [UUID.String] output instead.
//
// %v, %s, and %q print [UUID.String], honoring width, precision, and flags;
// %#v prints a Go composite literal. Any other verb formats the underlying
// [16]byte.
func (u UUID) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v', 's':
		if verb == 'v' && f.Flag('#') {
			u.formatGoSyntax(f)
			return
		}
		// Write the text directly rather than through a second Fprintf,
		// which would allocate a string and a format directive.
		var buf [36]byte
		encodeHex(buf[:], u)
		writePadded(f, buf[:])
	case 'q':
		_, _ = fmt.Fprintf(f, fmt.FormatString(f, verb), u.String())
	default:
		_, _ = fmt.Fprintf(f, fmt.FormatString(f, verb), [16]byte(u))
	}
}

// writePadded writes text the way fmt writes a string operand: truncated to
// the precision, then padded to the width with leading spaces, leading zeros
// under the '0' flag, or trailing spaces under the '-' flag. text must be
// ASCII, so that bytes and runes coincide.
func writePadded(f fmt.State, text []byte) {
	if p, ok := f.Precision(); ok && p < len(text) {
		text = text[:p]
	}
	w, _ := f.Width()
	pad := max(w-len(text), 0)
	if f.Flag('-') {
		_, _ = f.Write(text)
		writeRepeated(f, spaces, pad)
		return
	}
	fill := spaces
	if f.Flag('0') {
		fill = zeros
	}
	writeRepeated(f, fill, pad)
	_, _ = f.Write(text)
}

const (
	spaces = "                "
	zeros  = "0000000000000000"
)

// writeRepeated writes n bytes of fill (spaces or zeros) to f. Writing
// slices of a constant keeps padding allocation-free.
func writeRepeated(f fmt.State, fill string, n int) {
	for n > 0 {
		k := min(n, len(fill))
		_, _ = io.WriteString(f, fill[:k])
		n -= k
	}
}

// formatGoSyntax writes u as a Go composite literal, the way %#v prints any
// named byte array: uuid.UUID{0x6b, 0xa7, ...}.
func (u UUID) formatGoSyntax(f fmt.State) {
	_, _ = io.WriteString(f, "uuid.UUID{")
	for i, b := range u {
		if i > 0 {
			_, _ = io.WriteString(f, ", ")
		}
		_, _ = fmt.Fprintf(f, "%#x", b)
	}
	_, _ = io.WriteString(f, "}")
}

// URN returns the UUID in URN form: urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (u UUID) URN() string {
	var buf [45]byte
	copy(buf[:9], "urn:uuid:")
	encodeHex(buf[9:], u)
	return string(buf[:])
}

// AppendText appends the textual (36-char hyphenated) representation of u to b.
// It implements [encoding.TextAppender].
func (u UUID) AppendText(b []byte) ([]byte, error) {
	b = slices.Grow(b, 36)
	b = b[:len(b)+36]
	encodeHex(b[len(b)-36:], u)
	return b, nil
}

// AppendBinary appends the raw 16-byte representation of u to b.
// It implements [encoding.BinaryAppender].
func (u UUID) AppendBinary(b []byte) ([]byte, error) {
	return append(b, u[:]...), nil
}

// MarshalText returns the 36-character hyphenated representation.
// It implements [encoding.TextMarshaler].
// JSON encoding uses this method automatically.
func (u UUID) MarshalText() ([]byte, error) {
	var buf [36]byte
	encodeHex(buf[:], u)
	return buf[:], nil
}

// UnmarshalText parses a UUID from text in the strict 36-character form, as
// [Parse] does. It implements [encoding.TextUnmarshaler], so this is the rule
// JSON, XML, and other text encodings apply; URN, braced, and compact input
// is rejected. To accept those, see the LenientJSON example. On error, u is
// left unchanged.
func (u *UUID) UnmarshalText(data []byte) error {
	if len(data) != 36 {
		return &ParseError{Input: errInputBytes(data), Msg: "expected 36-character hyphenated format"}
	}
	if data[8] != '-' || data[13] != '-' || data[18] != '-' || data[23] != '-' {
		return hyphenError(string(data), 0)
	}
	// Decode into a local so a failure part-way through cannot leave *u
	// half-overwritten.
	var id UUID
	for i, x := range hexOffsets {
		v, ok := xtob(data[x], data[x+1])
		if !ok {
			return hexError(string(data), x)
		}
		id[i] = v
	}
	*u = id
	return nil
}

// MarshalBinary returns the raw 16-byte representation.
// It implements [encoding.BinaryMarshaler].
func (u UUID) MarshalBinary() ([]byte, error) {
	return u.Bytes(), nil
}

// UnmarshalBinary sets u from a 16-byte slice, as [FromBytes] does.
// It implements [encoding.BinaryUnmarshaler]. On error, u is left unchanged.
func (u *UUID) UnmarshalBinary(data []byte) error {
	id, err := FromBytes(data)
	if err != nil {
		return err
	}
	*u = id
	return nil
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

// Scan implements [database/sql.Scanner]. Unlike [UUID.UnmarshalText], it
// is lenient, because databases and drivers return UUIDs in different forms.
// It supports scanning from:
//   - string: text form parsed with [ParseLenient]
//   - []byte: 16 raw bytes (a BINARY(16) column), otherwise text form
//     parsed with [ParseLenient]
//
// A string is always parsed as text, so a 16-character value from a text
// column is rejected rather than silently read as raw bytes. Drivers deliver
// binary columns as []byte, which is the only source of raw bytes.
//
// Scanning SQL NULL is an error; use *UUID (nil pointer = NULL) instead.
// On error, u is left unchanged.
func (u *UUID) Scan(src any) error {
	var text string
	switch v := src.(type) {
	case string:
		text = v
	case []byte:
		if len(v) == 16 {
			*u = UUID(v)
			return nil
		}
		text = string(v)
	case nil:
		return errors.New("uuid: cannot scan NULL into UUID; use *UUID for nullable columns")
	default:
		return fmt.Errorf("uuid: cannot scan %T into UUID", src)
	}

	parsed, err := ParseLenient(text)
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

// Value implements [database/sql/driver.Valuer].
// It returns the UUID as a 36-character string, which suits native uuid
// column types (PostgreSQL, CockroachDB, MariaDB 10.7+) and text columns
// (SQLite has no UUID type). For BINARY(16) columns,
// wrap the type and return the raw bytes instead; see the
// [UUID.Value] example. [UUID.Scan] already accepts 16 raw bytes as []byte.
func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}
