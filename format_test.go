package uuid

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"strings"
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

func TestUnmarshalTextInputTruncated(t *testing.T) {
	long := bytes.Repeat([]byte("a"), 1000)
	var u UUID
	err := u.UnmarshalText(long)
	perr, ok := errors.AsType[*ParseError](err)
	if !ok {
		t.Fatalf("error type = %T, want *ParseError", err)
	}
	want := string(long[:64]) + "..."
	if perr.Input != want {
		t.Errorf("ParseError.Input = %q (len %d), want %q", perr.Input, len(perr.Input), want)
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

func TestScanString16CharsIsText(t *testing.T) {
	// A string is always text: a 16-character value from a text column must
	// be rejected, not silently read as the raw bytes of a UUID.
	for _, in := range []string{
		"not a uuid at al",
		string(MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8").Bytes()),
	} {
		var u UUID
		err := u.Scan(in)
		if _, ok := errors.AsType[*ParseError](err); !ok {
			t.Errorf("Scan(%q) error = %v, want *ParseError", in, err)
		}
		if !u.IsNil() {
			t.Errorf("Scan(%q) set u = %s, want it unchanged", in, u)
		}
	}
}

func TestDecodeErrorLeavesReceiverUnchanged(t *testing.T) {
	bad := []string{
		"short",
		"6ba7b810+9dad-11d1-80b4-00c04fd430c8",
		// Valid until the final digit, so a decoder writing in place would
		// already have overwritten 15 bytes.
		"00000000-0000-0000-0000-00000000000z",
	}
	decoders := map[string]func(*UUID, string) error{
		"UnmarshalText":   func(u *UUID, s string) error { return u.UnmarshalText([]byte(s)) },
		"Scan(string)":    func(u *UUID, s string) error { return u.Scan(s) },
		"Scan([]byte)":    func(u *UUID, s string) error { return u.Scan([]byte(s)) },
		"UnmarshalBinary": func(u *UUID, s string) error { return u.UnmarshalBinary([]byte(s)) },
	}
	for name, decode := range decoders {
		t.Run(name, func(t *testing.T) {
			for _, in := range bad {
				u := Max
				if err := decode(&u, in); err == nil {
					t.Fatalf("%s(%q) succeeded, want error", name, in)
				}
				if u != Max {
					t.Errorf("%s(%q) failed but changed u to %s", name, in, u)
				}
			}
		})
	}
}

func TestTextZeroAlloc(t *testing.T) {
	const s = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	data := []byte(s)
	u := MustParse(s)
	buf := make([]byte, 0, 64)
	for name, fn := range map[string]func(){
		"Parse":                func() { u, _ = Parse(s) },
		"ParseLenient":         func() { u, _ = ParseLenient(s) },
		"ParseLenient URN":     func() { u, _ = ParseLenient("urn:uuid:" + s) },
		"ParseLenient braced":  func() { u, _ = ParseLenient("{" + s + "}") },
		"ParseLenient compact": func() { u, _ = ParseLenient("6ba7b8109dad11d180b400c04fd430c8") },
		"UnmarshalText":        func() { _ = u.UnmarshalText(data) },
		"AppendText":           func() { buf, _ = u.AppendText(buf[:0]) },
	} {
		t.Run(name, func(t *testing.T) {
			if allocs := testing.AllocsPerRun(100, fn); allocs != 0 {
				t.Errorf("%s allocs = %v, want 0", name, allocs)
			}
		})
	}
}

func TestScanNil(t *testing.T) {
	var u UUID
	err := u.Scan(nil)
	if err == nil {
		t.Fatal("Scan(nil) should return error")
	}
	if !strings.Contains(err.Error(), "*UUID") {
		t.Errorf("Scan(nil) error should hint at *UUID, got: %v", err)
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

func TestJSONv2RoundTrip(t *testing.T) {
	type doc struct {
		ID     UUID  `json:"id"`
		Parent *UUID `json:"parent"`
	}
	want := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	b, err := jsonv2.Marshal(doc{ID: want})
	if err != nil {
		t.Fatalf("jsonv2.Marshal error: %v", err)
	}
	const wantJSON = `{"id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","parent":null}`
	if string(b) != wantJSON {
		t.Errorf("jsonv2.Marshal = %s, want %s", b, wantJSON)
	}

	var got doc
	if err := jsonv2.Unmarshal(b, &got); err != nil {
		t.Fatalf("jsonv2.Unmarshal error: %v", err)
	}
	if got.ID != want || got.Parent != nil {
		t.Errorf("jsonv2 round-trip = %+v, want ID %s and nil Parent", got, want)
	}

	err = jsonv2.Unmarshal([]byte(`{"id":"not-a-uuid"}`), &got)
	if _, ok := errors.AsType[*ParseError](err); !ok {
		t.Errorf("jsonv2.Unmarshal invalid: error = %T (%v), want *ParseError", err, err)
	}
}

// stringerOnly formats like UUID did before it implemented fmt.Formatter:
// fmt used its String method for string verbs and the [16]byte otherwise.
type stringerOnly [16]byte

func (s stringerOnly) String() string { return UUID(s).String() }

func TestFormat(t *testing.T) {
	for _, u := range []UUID{Nil, Max, MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")} {
		// Every verb except the hex ones prints exactly what it did when
		// fmt only saw the String method.
		for _, d := range []string{"%v", "%+v", "%s", "%q", "%d", "%40s|", "%-40s|", "%.8s", "%10.4v|", "%+q", "%#q", "%08b",
			"%040s|", "%-040s|", "%05.2s|", "%.0s|", "%.40s|", "%3s|", "%+40v|", "% 40s|", "%#40s|", "%10q|"} {
			if got, want := fmt.Sprintf(d, u), fmt.Sprintf(d, stringerOnly(u)); got != want {
				t.Errorf("Sprintf(%q, %s) = %q, want %q", d, u, got, want)
			}
		}
		oldGo := strings.Replace(fmt.Sprintf("%#v", stringerOnly(u)), "uuid.stringerOnly", "uuid.UUID", 1)
		if got := fmt.Sprintf("%#v", u); got != oldGo {
			t.Errorf("Sprintf(%%#v) = %q, want %q", got, oldGo)
		}
		if got, want := fmt.Sprint("id=", u), "id="+u.String(); got != want {
			t.Errorf("Sprint = %q, want %q", got, want)
		}

		// The hex verbs format the 16 bytes, not the 36-character text.
		for _, d := range []string{"%x", "%X", "% x", "%#x", "%40x|", "%-40X|"} {
			if got, want := fmt.Sprintf(d, u), fmt.Sprintf(d, [16]byte(u)); got != want {
				t.Errorf("Sprintf(%q, %s) = %q, want %q", d, u, got, want)
			}
		}
	}
	if got, want := fmt.Sprintf("%x", MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")), "6ba7b8109dad11d180b400c04fd430c8"; got != want {
		t.Errorf("%%x = %q, want %q", got, want)
	}
	// Containers format their elements through Format too.
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if got, want := fmt.Sprintf("%v", []UUID{u, Nil}), "["+u.String()+" "+Nil.String()+"]"; got != want {
		t.Errorf("slice %%v = %q, want %q", got, want)
	}
}
