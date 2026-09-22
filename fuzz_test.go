package uuid

import (
	"strings"
	"testing"
)

func FuzzParse(f *testing.F) {
	// Seed corpus with valid and interesting inputs
	f.Add("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	f.Add("00000000-0000-0000-0000-000000000000")
	f.Add("ffffffff-ffff-ffff-ffff-ffffffffffff")
	f.Add("550e8400-e29b-41d4-a716-446655440000")
	f.Add("")
	f.Add("not-a-uuid")
	f.Add("FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF")

	f.Fuzz(func(t *testing.T, s string) {
		u, err := Parse(s)
		if err != nil {
			return
		}
		// If parse succeeded, round-trip must be exact
		got := u.String()
		u2, err := Parse(got)
		if err != nil {
			t.Fatalf("round-trip Parse failed: %v", err)
		}
		if u != u2 {
			t.Fatalf("round-trip mismatch: %v != %v", u, u2)
		}
	})
}

// FuzzDecodersAgree cross-checks the separate hand-written decoders: Parse,
// UnmarshalText, ParseLenient (all four forms), and Scan must accept the
// same inputs, produce the same UUID, and fail with the same error.
func FuzzDecodersAgree(f *testing.F) {
	f.Add("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	f.Add("6BA7B810-9DAD-11D1-80B4-00C04FD430C8")
	f.Add("6ba7b810-9dad-11d1-80b4-00c04fd430cg")
	f.Add("6ba7b810+9dad-11d1-80b4-00c04fd430c8")
	f.Add("urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	f.Add("6ba7b8109dad11d180b400c04fd430c8")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		u, err := Parse(s)

		var fromText UUID
		textErr := fromText.UnmarshalText([]byte(s))
		if (err == nil) != (textErr == nil) {
			t.Fatalf("Parse err = %v, UnmarshalText err = %v", err, textErr)
		}
		if err != nil {
			if err.Error() != textErr.Error() {
				t.Fatalf("error mismatch:\n Parse:         %v\n UnmarshalText: %v", err, textErr)
			}
			return
		}
		if fromText != u {
			t.Fatalf("UnmarshalText = %s, Parse = %s", fromText, u)
		}
		if u2, err := Parse(u.String()); err != nil || u2 != u {
			t.Fatalf("round-trip Parse(%q) = %s, %v", u.String(), u2, err)
		}

		// Every form ParseLenient accepts must decode to the same UUID.
		compact := strings.ReplaceAll(s, "-", "")
		for _, form := range []string{s, "urn:uuid:" + s, "URN:UUID:" + s, "{" + s + "}", compact} {
			if got, err := ParseLenient(form); err != nil || got != u {
				t.Fatalf("ParseLenient(%q) = %s, %v; want %s", form, got, err, u)
			}
			var scanned UUID
			if err := scanned.Scan(form); err != nil || scanned != u {
				t.Fatalf("Scan(%q) = %s, %v; want %s", form, scanned, err, u)
			}
		}
	})
}

func FuzzParseLenient(f *testing.F) {
	f.Add("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	f.Add("urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	f.Add("{6ba7b810-9dad-11d1-80b4-00c04fd430c8}")
	f.Add("6ba7b8109dad11d180b400c04fd430c8")
	f.Add("")
	f.Add("not-a-uuid")

	f.Fuzz(func(t *testing.T, s string) {
		u, err := ParseLenient(s)
		if err != nil {
			return
		}
		// If parse succeeded, strict round-trip must work
		got := u.String()
		u2, err := Parse(got)
		if err != nil {
			t.Fatalf("round-trip Parse failed after ParseLenient: %v", err)
		}
		if u != u2 {
			t.Fatalf("round-trip mismatch: %v != %v", u, u2)
		}
	})
}
