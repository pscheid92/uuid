package uuid

import (
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
