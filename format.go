package uuid

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
