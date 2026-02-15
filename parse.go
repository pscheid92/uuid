package uuid

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
	return parseHex(s, 0)
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
	// groups: 8-4-4-4-12 hex digits
	// byte positions in UUID: 0-3, 4-5, 6-7, 8-9, 10-15
	src := offset
	for i := range 16 {
		// skip hyphens
		if src-offset == 8 || src-offset == 13 || src-offset == 18 || src-offset == 23 {
			src++
		}
		hi, ok1 := fromHexChar(s[src])
		lo, ok2 := fromHexChar(s[src+1])
		if !ok1 || !ok2 {
			return Nil, &ParseError{Input: s, Msg: "invalid hex character"}
		}
		u[i] = hi<<4 | lo
		src += 2
	}
	return u, nil
}

// parseCompact decodes a 32-character hex string with no hyphens.
func parseCompact(s string) (UUID, error) {
	var u UUID
	for i := range 16 {
		hi, ok1 := fromHexChar(s[i*2])
		lo, ok2 := fromHexChar(s[i*2+1])
		if !ok1 || !ok2 {
			return Nil, &ParseError{Input: s, Msg: "invalid hex character"}
		}
		u[i] = hi<<4 | lo
	}
	return u, nil
}

// fromHexChar converts a hex character to its value.
func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}
