package uuid

import "testing"

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
