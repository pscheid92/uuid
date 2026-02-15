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
		{"00000000-0000-3000-8000-000000000000", Version3},
		{"00000000-0000-4000-8000-000000000000", Version4},
		{"00000000-0000-5000-8000-000000000000", Version5},
		{"00000000-0000-7000-8000-000000000000", Version7},
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

	got := u.Time()
	if !got.Equal(now) {
		t.Errorf("Time() = %v, want %v", got, now)
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
		{VersionNil, "NIL"},
		{Version3, "V3"},
		{Version4, "V4"},
		{Version5, "V5"},
		{Version7, "V7"},
		{Version8, "V8"},
		{VersionMax, "MAX"},
		{Version(2), "unknown"},
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
