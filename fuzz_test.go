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
// UnmarshalText, ParseLenient (all four forms), and Scan (string and []byte)
// must accept the same inputs, produce the same UUID, and fail with the same
// error.
func FuzzDecodersAgree(f *testing.F) {
	f.Add("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	f.Add("6BA7B810-9DAD-11D1-80B4-00C04FD430C8")
	f.Add("6ba7b810-9dad-11d1-80b4-00c04fd430cg")
	f.Add("6ba7b810+9dad-11d1-80b4-00c04fd430c8")
	f.Add("urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	f.Add("{6ba7b810-9dad-11d1-80b4-00c04fd430c8}")
	f.Add("6ba7b8109dad11d180b400c04fd430c8")
	f.Add("6ba7b8109dad11d180b400c04fd430cg")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		u, err := Parse(s)

		var fromText UUID
		textErr := fromText.UnmarshalText([]byte(s))
		sameResult(t, "UnmarshalText", fromText, textErr, u, err)

		// ParseLenient's standard form shares Parse's contract, errors included.
		lenient, lenientErr := ParseLenient(s)
		if len(s) == 36 {
			sameResult(t, "ParseLenient", lenient, lenientErr, u, err)
		}

		// Scan parses text with ParseLenient, whether it arrives as a string
		// or as a []byte (except a 16-byte []byte, which is raw).
		var fromString UUID
		stringErr := fromString.Scan(s)
		sameResult(t, "Scan(string)", fromString, stringErr, lenient, lenientErr)
		if len(s) != 16 {
			var fromBytes UUID
			bytesErr := fromBytes.Scan([]byte(s))
			sameResult(t, "Scan([]byte)", fromBytes, bytesErr, lenient, lenientErr)
		}

		if err != nil {
			return
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
			var fromString, fromBytes UUID
			if err := fromString.Scan(form); err != nil || fromString != u {
				t.Fatalf("Scan(%q) = %s, %v; want %s", form, fromString, err, u)
			}
			if err := fromBytes.Scan([]byte(form)); err != nil || fromBytes != u {
				t.Fatalf("Scan([]byte(%q)) = %s, %v; want %s", form, fromBytes, err, u)
			}
		}
	})
}

// sameResult fails t unless a decoder returned the same UUID and the same
// error (or lack of one) as the reference decoder.
func sameResult(t *testing.T, name string, got UUID, gotErr error, want UUID, wantErr error) {
	t.Helper()
	if got != want || (gotErr == nil) != (wantErr == nil) || (gotErr != nil && gotErr.Error() != wantErr.Error()) {
		t.Fatalf("%s = %s, %v; want %s, %v", name, got, gotErr, want, wantErr)
	}
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
