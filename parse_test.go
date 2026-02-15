package uuid

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c8", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{"00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000"},
		{"ffffffff-ffff-ffff-ffff-ffffffffffff", "ffffffff-ffff-ffff-ffff-ffffffffffff"},
		{"FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF", "ffffffff-ffff-ffff-ffff-ffffffffffff"},
		{"550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440000"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			u, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}
			if got := u.String(); got != tt.want {
				t.Errorf("Parse(%q) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"", "empty"},
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c", "too short"},
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c8a", "too long"},
		{"6ba7b810+9dad-11d1-80b4-00c04fd430c8", "wrong separator"},
		{"6ba7b810-9dad+11d1-80b4-00c04fd430c8", "wrong separator 2"},
		{"6ba7b810-9dad-11d1+80b4-00c04fd430c8", "wrong separator 3"},
		{"6ba7b810-9dad-11d1-80b4+00c04fd430c8", "wrong separator 4"},
		{"6ba7b810-9dad-11d1-80b4-00c04fd430cg", "invalid hex"},
		{"urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8", "URN not accepted"},
		{"{6ba7b810-9dad-11d1-80b4-00c04fd430c8}", "braced not accepted"},
		{"6ba7b8109dad11d180b400c04fd430c8", "compact not accepted"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Fatalf("Parse(%q) should return error", tt.input)
			}
			var perr *ParseError
			if !errors.As(err, &perr) {
				t.Fatalf("Parse(%q) error type = %T, want *ParseError", tt.input, err)
			}
		})
	}
}

func TestParseErrorsAsType(t *testing.T) {
	_, err := Parse("not-a-uuid")
	perr, ok := errors.AsType[*ParseError](err)
	if !ok {
		t.Fatalf("errors.AsType[*ParseError] returned false")
	}
	if perr.Input != "not-a-uuid" {
		t.Errorf("ParseError.Input = %q, want %q", perr.Input, "not-a-uuid")
	}
}

func TestParseLenient(t *testing.T) {
	want := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	tests := []struct {
		name  string
		input string
	}{
		{"standard", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{"URN", "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{"braced", "{6ba7b810-9dad-11d1-80b4-00c04fd430c8}"},
		{"compact", "6ba7b8109dad11d180b400c04fd430c8"},
		{"compact upper", "6BA7B8109DAD11D180B400C04FD430C8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := ParseLenient(tt.input)
			if err != nil {
				t.Fatalf("ParseLenient(%q) unexpected error: %v", tt.input, err)
			}
			if got := u.String(); got != want {
				t.Errorf("ParseLenient(%q) = %s, want %s", tt.input, got, want)
			}
		})
	}
}

func TestParseLenientErrors(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"", "empty"},
		{"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", "invalid hex"},
		{"abc:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8", "wrong URN prefix"},
		{"[6ba7b810-9dad-11d1-80b4-00c04fd430c8]", "wrong braces"},
		{"6ba7b8109dad11d180b400c04fd430cg", "invalid hex compact"},
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c8-extra", "too long"},
		{"short", "too short"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, err := ParseLenient(tt.input)
			if err == nil {
				t.Fatalf("ParseLenient(%q) should return error", tt.input)
			}
		})
	}
}

func TestParseLenientURNBadHyphens(t *testing.T) {
	_, err := ParseLenient("urn:uuid:6ba7b810+9dad-11d1-80b4-00c04fd430c8")
	if err == nil {
		t.Fatal("expected error for URN with bad hyphens")
	}
}

func TestParseLenientBracedBadHyphens(t *testing.T) {
	_, err := ParseLenient("{6ba7b810+9dad-11d1-80b4-00c04fd430c8}")
	if err == nil {
		t.Fatal("expected error for braced with bad hyphens")
	}
}

func TestMustParse(t *testing.T) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if u.String() != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" {
		t.Errorf("MustParse returned wrong UUID")
	}
}

func TestMustParsePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustParse should panic on invalid input")
		}
	}()
	MustParse("invalid")
}

func TestFromBytes(t *testing.T) {
	want := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	b := want.Bytes()
	got, err := FromBytes(b)
	if err != nil {
		t.Fatalf("FromBytes() unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("FromBytes() = %v, want %v", got, want)
	}
}

func TestFromBytesError(t *testing.T) {
	_, err := FromBytes([]byte{1, 2, 3})
	if err == nil {
		t.Fatal("FromBytes should fail for wrong length")
	}
	var lerr *LengthError
	if !errors.As(err, &lerr) {
		t.Fatalf("error type = %T, want *LengthError", err)
	}
	if lerr.Got != 3 {
		t.Errorf("LengthError.Got = %d, want 3", lerr.Got)
	}

	// Also test errors.AsType
	lerr2, ok := errors.AsType[*LengthError](err)
	if !ok {
		t.Fatal("errors.AsType[*LengthError] returned false")
	}
	if lerr2.Got != 3 {
		t.Errorf("errors.AsType LengthError.Got = %d, want 3", lerr2.Got)
	}
}

func TestParseRoundTrip(t *testing.T) {
	inputs := []string{
		"00000000-0000-0000-0000-000000000000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"550e8400-e29b-41d4-a716-446655440000",
	}
	for _, s := range inputs {
		u, err := Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q): %v", s, err)
		}
		if got := u.String(); got != s {
			t.Errorf("round-trip: Parse(%q).String() = %q", s, got)
		}
	}
}
