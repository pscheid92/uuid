package uuid

import (
	"testing"
	"time"
)

func TestUUIDZeroValue(t *testing.T) {
	var u UUID
	if u != Nil {
		t.Errorf("zero value should equal Nil")
	}
	if !u.IsNil() {
		t.Errorf("zero value IsNil() should be true")
	}
}

func TestMax(t *testing.T) {
	for i, b := range Max {
		if b != 0xff {
			t.Errorf("Max[%d] = %#x, want 0xff", i, b)
		}
	}
}

func TestNamespaceConstants(t *testing.T) {
	tests := []struct {
		name string
		uuid UUID
		want string
	}{
		{"DNS", NamespaceDNS, "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{"URL", NamespaceURL, "6ba7b811-9dad-11d1-80b4-00c04fd430c8"},
		{"OID", NamespaceOID, "6ba7b812-9dad-11d1-80b4-00c04fd430c8"},
		{"X500", NamespaceX500, "6ba7b814-9dad-11d1-80b4-00c04fd430c8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.uuid.String(); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	tests := []struct {
		hex     string
		version Version
	}{
		{"00000000-0000-4000-8000-000000000000", V4},
		{"00000000-0000-5000-8000-000000000000", V5},
		{"00000000-0000-7000-8000-000000000000", V7},
		{"00000000-0000-8000-8000-000000000000", V8},
	}
	for _, tt := range tests {
		u := MustParse(tt.hex)
		if got := u.Version(); got != tt.version {
			t.Errorf("Parse(%q).Version() = %v, want %v", tt.hex, got, tt.version)
		}
	}
}

func TestVariant(t *testing.T) {
	tests := []struct {
		name    string
		byte8   byte
		variant Variant
	}{
		{"NCS", 0x00, VariantNCS},
		{"NCS upper", 0x7f, VariantNCS},
		{"RFC9562", 0x80, VariantRFC9562},
		{"RFC9562 upper", 0xbf, VariantRFC9562},
		{"Microsoft", 0xc0, VariantMicrosoft},
		{"Microsoft upper", 0xdf, VariantMicrosoft},
		{"Future", 0xe0, VariantFuture},
		{"Future upper", 0xff, VariantFuture},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u UUID
			u[8] = tt.byte8
			if got := u.Variant(); got != tt.variant {
				t.Errorf("variant byte %#x: got %v, want %v", tt.byte8, got, tt.variant)
			}
		})
	}
}

func TestString(t *testing.T) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	want := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	if got := u.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestURN(t *testing.T) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	want := "urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	if got := u.URN(); got != want {
		t.Errorf("URN() = %q, want %q", got, want)
	}
}

func TestBytes(t *testing.T) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	b := u.Bytes()
	if len(b) != 16 {
		t.Fatalf("Bytes() length = %d, want 16", len(b))
	}
	// Verify it's a copy
	b[0] = 0xff
	if u[0] == 0xff {
		t.Errorf("Bytes() should return a copy, not a reference")
	}
}

func TestIsNil(t *testing.T) {
	if !Nil.IsNil() {
		t.Errorf("Nil.IsNil() should be true")
	}
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if u.IsNil() {
		t.Errorf("non-nil UUID.IsNil() should be false")
	}
}

func TestCompare(t *testing.T) {
	a := MustParse("00000000-0000-0000-0000-000000000001")
	b := MustParse("00000000-0000-0000-0000-000000000002")

	if Compare(a, b) != -1 {
		t.Errorf("Compare(a, b) should be -1")
	}
	if Compare(b, a) != 1 {
		t.Errorf("Compare(b, a) should be 1")
	}
	if Compare(a, a) != 0 {
		t.Errorf("Compare(a, a) should be 0")
	}

	// The method matches the package-level function.
	for _, pair := range [][2]UUID{{a, b}, {b, a}, {a, a}, {Nil, Max}} {
		if got, want := pair[0].Compare(pair[1]), Compare(pair[0], pair[1]); got != want {
			t.Errorf("%s.Compare(%s) = %d, want %d", pair[0], pair[1], got, want)
		}
	}
}

func TestTimeNonRFCVariant(t *testing.T) {
	// Version bits read 7, but the version field is only defined for the
	// RFC 9562 variant, so none of these carry a V7 timestamp.
	v7 := NewV7At(time.Unix(1_700_000_000, 0))
	for _, tc := range []struct {
		name string
		b8   byte
	}{
		{"NCS", 0x00},
		{"Microsoft", 0xc0},
		{"Future", 0xe0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			u := v7
			u[8] = tc.b8 | u[8]&0x1f
			if u.Variant().String() != tc.name {
				t.Fatalf("test setup: Variant() = %v, want %s", u.Variant(), tc.name)
			}
			if got, ok := u.Time(); ok || !got.IsZero() {
				t.Errorf("Time() = (%v, %v), want (zero, false)", got, ok)
			}
		})
	}
}

func TestTimeV7(t *testing.T) {
	// Build a V7 UUID with a known timestamp
	now := time.Now().Truncate(time.Millisecond)
	ms := now.UnixMilli()

	var u UUID
	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)
	u[6] = 0x70 // version 7
	u[8] = 0x80 // variant RFC9562

	got, ok := u.Time()
	if !ok {
		t.Fatal("Time() ok = false for a V7 UUID")
	}
	if !got.Equal(now) {
		t.Errorf("Time() = %v, want %v", got, now)
	}
}

func TestUUIDTimeNonV7(t *testing.T) {
	// A V1 UUID carries a timestamp in a different layout; V4 carries none.
	// Both must report ok=false rather than a plausible but wrong time.
	for _, in := range []string{
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8", // V1 (NamespaceDNS)
		"550e8400-e29b-41d4-a716-446655440000", // V4
		"00000000-0000-0000-0000-000000000000", // Nil
		"ffffffff-ffff-ffff-ffff-ffffffffffff", // Max
	} {
		u := MustParse(in)
		got, ok := u.Time()
		if ok {
			t.Errorf("Time() ok = true for %s (%s)", in, u.Version())
		}
		if !got.IsZero() {
			t.Errorf("Time() = %v for %s, want zero time", got, in)
		}
	}
}

func TestUUIDComparable(t *testing.T) {
	// Verify UUID can be used as a map key
	a := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	b := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	m := map[UUID]bool{a: true}
	if !m[b] {
		t.Errorf("UUID should be usable as a map key")
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		v    Version
		want string
	}{
		{VNil, "NIL"},
		{V4, "V4"},
		{V5, "V5"},
		{V7, "V7"},
		{V8, "V8"},
		{VMax, "MAX"},
		{Version(1), "V1"},
		{Version(2), "V2"},
		{Version(3), "V3"},
		{Version(6), "V6"},
		{Version(9), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("Version(%d).String() = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestVariantString(t *testing.T) {
	tests := []struct {
		v    Variant
		want string
	}{
		{VariantNCS, "NCS"},
		{VariantRFC9562, "RFC9562"},
		{VariantMicrosoft, "Microsoft"},
		{VariantFuture, "Future"},
		{Variant(42), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("Variant(%d).String() = %q, want %q", tt.v, got, tt.want)
		}
	}
}
