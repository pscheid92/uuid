package uuid

import (
	"database/sql/driver"
	"testing"
)

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

func TestValueInterface(t *testing.T) {
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
