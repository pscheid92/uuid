package uuid

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
)

func TestAppendText(t *testing.T) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	buf, err := u.AppendText(nil)
	if err != nil {
		t.Fatalf("AppendText() error: %v", err)
	}
	if string(buf) != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Errorf("AppendText() = %q", buf)
	}

	// Append to existing data
	prefix := []byte("uuid:")
	buf, err = u.AppendText(prefix)
	if err != nil {
		t.Fatalf("AppendText(prefix) error: %v", err)
	}
	if string(buf) != "uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Errorf("AppendText(prefix) = %q", buf)
	}

	// Force grow reallocation: full-capacity slice with no room for 36 bytes
	tight := make([]byte, 4)
	copy(tight, "pre:")
	buf, err = u.AppendText(tight)
	if err != nil {
		t.Fatalf("AppendText(tight) error: %v", err)
	}
	if string(buf) != "pre:6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Errorf("AppendText(tight) = %q", buf)
	}

	// Exercise grow fast path: slice with plenty of spare capacity
	spacious := make([]byte, 4, 50)
	copy(spacious, "pre:")
	buf, err = u.AppendText(spacious)
	if err != nil {
		t.Fatalf("AppendText(spacious) error: %v", err)
	}
	if string(buf) != "pre:6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Errorf("AppendText(spacious) = %q", buf)
	}
}

func TestAppendBinary(t *testing.T) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	buf, err := u.AppendBinary(nil)
	if err != nil {
		t.Fatalf("AppendBinary() error: %v", err)
	}
	if len(buf) != 16 {
		t.Errorf("AppendBinary() length = %d, want 16", len(buf))
	}
	got, err := FromBytes(buf)
	if err != nil {
		t.Fatalf("FromBytes() error: %v", err)
	}
	if got != u {
		t.Errorf("AppendBinary round-trip failed")
	}
}

func TestMarshalText(t *testing.T) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	b, err := u.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error: %v", err)
	}
	if string(b) != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Errorf("MarshalText() = %q", b)
	}
}

func TestUnmarshalText(t *testing.T) {
	var u UUID
	err := u.UnmarshalText([]byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	if err != nil {
		t.Fatalf("UnmarshalText() error: %v", err)
	}
	want := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if u != want {
		t.Errorf("UnmarshalText() = %v, want %v", u, want)
	}
}

func TestUnmarshalTextError(t *testing.T) {
	var u UUID
	err := u.UnmarshalText([]byte("invalid"))
	if err == nil {
		t.Fatal("UnmarshalText should fail on invalid input")
	}
}

func TestUnmarshalTextBadHyphens(t *testing.T) {
	var u UUID
	err := u.UnmarshalText([]byte("6ba7b810+9dad-11d1-80b4-00c04fd430c8"))
	if err == nil {
		t.Fatal("UnmarshalText should fail on bad hyphens")
	}
}

func TestUnmarshalTextInvalidHex(t *testing.T) {
	var u UUID
	// Valid length and hyphens, but 'zz' is invalid hex
	err := u.UnmarshalText([]byte("zza7b810-9dad-11d1-80b4-00c04fd430c8"))
	if err == nil {
		t.Fatal("UnmarshalText should fail on invalid hex")
	}
}

func TestMarshalBinary(t *testing.T) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	b, err := u.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error: %v", err)
	}
	if len(b) != 16 {
		t.Errorf("MarshalBinary() length = %d, want 16", len(b))
	}
}

func TestUnmarshalBinary(t *testing.T) {
	original := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	b, _ := original.MarshalBinary()
	var u UUID
	err := u.UnmarshalBinary(b)
	if err != nil {
		t.Fatalf("UnmarshalBinary() error: %v", err)
	}
	if u != original {
		t.Errorf("UnmarshalBinary() = %v, want %v", u, original)
	}
}

func TestUnmarshalBinaryError(t *testing.T) {
	var u UUID
	err := u.UnmarshalBinary([]byte{1, 2, 3})
	if err == nil {
		t.Fatal("UnmarshalBinary should fail on wrong length")
	}
	lerr, ok := errors.AsType[*LengthError](err)
	if !ok {
		t.Fatalf("error type = %T, want *LengthError", err)
	}
	if lerr.Got != 3 {
		t.Errorf("LengthError.Got = %d, want 3", lerr.Got)
	}
}

func TestMarshalTextRoundTrip(t *testing.T) {
	original := MustParse("550e8400-e29b-41d4-a716-446655440000")
	b, err := original.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error: %v", err)
	}
	var u UUID
	if err := u.UnmarshalText(b); err != nil {
		t.Fatalf("UnmarshalText() error: %v", err)
	}
	if u != original {
		t.Errorf("round-trip failed: got %v, want %v", u, original)
	}
}

func TestMarshalBinaryRoundTrip(t *testing.T) {
	original := MustParse("550e8400-e29b-41d4-a716-446655440000")
	b, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error: %v", err)
	}
	var u UUID
	if err := u.UnmarshalBinary(b); err != nil {
		t.Fatalf("UnmarshalBinary() error: %v", err)
	}
	if u != original {
		t.Errorf("round-trip failed: got %v, want %v", u, original)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	type doc struct {
		ID UUID `json:"id"`
	}
	original := doc{ID: MustParse("550e8400-e29b-41d4-a716-446655440000")}
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}
	if string(b) != `{"id":"550e8400-e29b-41d4-a716-446655440000"}` {
		t.Errorf("json.Marshal() = %s", b)
	}
	var decoded doc
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if decoded != original {
		t.Errorf("JSON round-trip failed: got %v, want %v", decoded, original)
	}
}

func TestJSONNull(t *testing.T) {
	type doc struct {
		ID *UUID `json:"id"`
	}
	// nil pointer -> JSON null
	original := doc{ID: nil}
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}
	if string(b) != `{"id":null}` {
		t.Errorf("json.Marshal() = %s, want null", b)
	}
	// JSON null -> nil pointer
	var decoded doc
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}
	if decoded.ID != nil {
		t.Errorf("expected nil ID after null unmarshal, got %v", decoded.ID)
	}
}

func TestScanString(t *testing.T) {
	var u UUID
	err := u.Scan("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		t.Fatalf("Scan(string) error: %v", err)
	}
	want := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if u != want {
		t.Errorf("Scan(string) = %v, want %v", u, want)
	}
}

func TestScanBytes16(t *testing.T) {
	want := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	raw := want.Bytes()
	var u UUID
	err := u.Scan(raw)
	if err != nil {
		t.Fatalf("Scan([]byte{16}) error: %v", err)
	}
	if u != want {
		t.Errorf("Scan([]byte{16}) = %v, want %v", u, want)
	}
}

func TestScanBytesText(t *testing.T) {
	var u UUID
	err := u.Scan([]byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	if err != nil {
		t.Fatalf("Scan([]byte text) error: %v", err)
	}
	want := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if u != want {
		t.Errorf("Scan([]byte text) = %v, want %v", u, want)
	}
}

func TestScanInvalidType(t *testing.T) {
	var u UUID
	err := u.Scan(42)
	if err == nil {
		t.Fatal("Scan(int) should return error")
	}
}

func TestScanInvalidString(t *testing.T) {
	var u UUID
	err := u.Scan("not-a-uuid")
	if err == nil {
		t.Fatal("Scan(invalid string) should return error")
	}
}

func TestScanInvalidBytes(t *testing.T) {
	var u UUID
	err := u.Scan([]byte{1, 2, 3})
	if err == nil {
		t.Fatal("Scan(short bytes) should return error")
	}
}

func TestValue(t *testing.T) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	v, err := u.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("Value() type = %T, want string", v)
	}
	if s != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Errorf("Value() = %q", s)
	}
}

func TestValueInterface(_ *testing.T) {
	// Verify UUID implements driver.Valuer
	var _ driver.Valuer = UUID{}
}

func TestScanLenientFormats(t *testing.T) {
	want := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	tests := []struct {
		name  string
		input string
	}{
		{"standard", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{"compact", "6ba7b8109dad11d180b400c04fd430c8"},
		{"URN", "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{"braced", "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u UUID
			if err := u.Scan(tt.input); err != nil {
				t.Fatalf("Scan(%q) error: %v", tt.input, err)
			}
			if u != want {
				t.Errorf("Scan(%q) = %v, want %v", tt.input, u, want)
			}
		})
	}
}

func TestScanValueRoundTrip(t *testing.T) {
	original := MustParse("550e8400-e29b-41d4-a716-446655440000")
	v, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	var decoded UUID
	if err := decoded.Scan(v); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if decoded != original {
		t.Errorf("round-trip failed: %v != %v", decoded, original)
	}
}
