package uuid

import (
	"database/sql/driver"
	"fmt"
)

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

// AppendText appends the textual (36-char hyphenated) representation of u to b.
// It implements [encoding.TextAppender].
func (u UUID) AppendText(b []byte) ([]byte, error) {
	b = grow(b, 36)
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

// UnmarshalText parses a UUID from text (strict 36-char format).
// It implements [encoding.TextUnmarshaler].
func (u *UUID) UnmarshalText(data []byte) error {
	if len(data) != 36 {
		return &ParseError{Input: string(data), Msg: "expected 36-character hyphenated format"}
	}
	if data[8] != '-' || data[13] != '-' || data[18] != '-' || data[23] != '-' {
		return &ParseError{Input: string(data), Msg: "expected hyphens at positions 8, 13, 18, 23"}
	}
	if !parseHexBytes(u, data, 0) {
		return &ParseError{Input: string(data), Msg: "invalid hex character"}
	}
	return nil
}

// MarshalBinary returns the raw 16-byte representation.
// It implements [encoding.BinaryMarshaler].
func (u UUID) MarshalBinary() ([]byte, error) {
	b := make([]byte, 16)
	copy(b, u[:])
	return b, nil
}

// UnmarshalBinary sets u from a 16-byte slice.
// It implements [encoding.BinaryUnmarshaler].
func (u *UUID) UnmarshalBinary(data []byte) error {
	if len(data) != 16 {
		return &LengthError{Got: len(data), Want: "16 bytes"}
	}
	copy(u[:], data)
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

// grow appends n zero bytes to b and returns the extended slice.
func grow(b []byte, n int) []byte {
	l := len(b)
	if cap(b)-l >= n {
		return b[:l+n]
	}
	newBuf := make([]byte, l+n, (l+n)*2)
	copy(newBuf, b)
	return newBuf
}

// Scan implements [database/sql.Scanner]. It supports scanning from:
//   - string: parsed with [ParseLenient]
//   - []byte: 16 raw bytes or text form parsed with [ParseLenient]
//
// For SQL NULL handling, use *UUID (nil pointer = NULL).
func (u *UUID) Scan(src any) error {
	switch v := src.(type) {
	case string:
		parsed, err := ParseLenient(v)
		if err != nil {
			return err
		}
		*u = parsed
		return nil

	case []byte:
		if len(v) == 16 {
			copy(u[:], v)
			return nil
		}
		parsed, err := ParseLenient(string(v))
		if err != nil {
			return err
		}
		*u = parsed
		return nil

	default:
		return fmt.Errorf("uuid: cannot scan %T into UUID", src)
	}
}

// Value implements [database/sql/driver.Valuer].
// It returns the UUID as a 36-character string.
func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}
