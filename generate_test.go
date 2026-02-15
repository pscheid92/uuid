package uuid

import (
	"testing"
	"testing/cryptotest"
)

func TestNewV4(t *testing.T) {
	cryptotest.SetGlobalRandom(t, 42)

	u := NewV4()
	if u.Version() != Version4 {
		t.Errorf("NewV4().Version() = %v, want V4", u.Version())
	}
	if u.Variant() != VariantRFC9562 {
		t.Errorf("NewV4().Variant() = %v, want RFC9562", u.Variant())
	}
	if u.IsNil() {
		t.Errorf("NewV4() should not be nil")
	}
}

func TestNewV4Deterministic(t *testing.T) {
	cryptotest.SetGlobalRandom(t, 123)
	a := NewV4()

	cryptotest.SetGlobalRandom(t, 123)
	b := NewV4()

	if a != b {
		t.Errorf("NewV4 with same seed should produce same UUID: %s != %s", a, b)
	}
}

func TestNewV4Uniqueness(t *testing.T) {
	seen := make(map[UUID]bool)
	for range 1000 {
		u := NewV4()
		if seen[u] {
			t.Fatalf("duplicate V4 UUID: %s", u)
		}
		seen[u] = true
	}
}

func TestNewV3(t *testing.T) {
	// RFC 9562 Appendix B.1 test vector
	u := NewV3(NamespaceDNS, "www.example.com")
	if u.Version() != Version3 {
		t.Errorf("NewV3().Version() = %v, want V3", u.Version())
	}
	if u.Variant() != VariantRFC9562 {
		t.Errorf("NewV3().Variant() = %v, want RFC9562", u.Variant())
	}
	want := MustParse("5df41881-3aed-3515-88a7-2f4a814cf09e")
	if u != want {
		t.Errorf("NewV3(DNS, www.example.com) = %s, want %s", u, want)
	}
}

func TestNewV3Deterministic(t *testing.T) {
	a := NewV3(NamespaceDNS, "test")
	b := NewV3(NamespaceDNS, "test")
	if a != b {
		t.Errorf("NewV3 should be deterministic: %s != %s", a, b)
	}
}

func TestNewV3CustomNamespace(t *testing.T) {
	ns := MustParse("12345678-1234-1234-1234-123456789abc")
	u := NewV3(ns, "hello")
	if u.Version() != Version3 {
		t.Errorf("Version = %v, want V3", u.Version())
	}
	if u.Variant() != VariantRFC9562 {
		t.Errorf("Variant = %v, want RFC9562", u.Variant())
	}
	// Same input must produce same output
	u2 := NewV3(ns, "hello")
	if u != u2 {
		t.Errorf("determinism failed")
	}
}

func TestNewV5(t *testing.T) {
	// RFC 9562 Appendix B.2 test vector
	u := NewV5(NamespaceDNS, "www.example.com")
	if u.Version() != Version5 {
		t.Errorf("NewV5().Version() = %v, want V5", u.Version())
	}
	if u.Variant() != VariantRFC9562 {
		t.Errorf("NewV5().Variant() = %v, want RFC9562", u.Variant())
	}
	want := MustParse("2ed6657d-e927-568b-95e1-2665a8aea6a2")
	if u != want {
		t.Errorf("NewV5(DNS, www.example.com) = %s, want %s", u, want)
	}
}

func TestNewV5Deterministic(t *testing.T) {
	a := NewV5(NamespaceURL, "https://example.com")
	b := NewV5(NamespaceURL, "https://example.com")
	if a != b {
		t.Errorf("NewV5 should be deterministic: %s != %s", a, b)
	}
}

func TestNewV5AllNamespaces(t *testing.T) {
	namespaces := []struct {
		name string
		ns   UUID
	}{
		{"DNS", NamespaceDNS},
		{"URL", NamespaceURL},
		{"OID", NamespaceOID},
		{"X500", NamespaceX500},
	}
	for _, tt := range namespaces {
		t.Run(tt.name, func(t *testing.T) {
			u := NewV5(tt.ns, "test")
			if u.Version() != Version5 {
				t.Errorf("Version = %v, want V5", u.Version())
			}
			if u.Variant() != VariantRFC9562 {
				t.Errorf("Variant = %v, want RFC9562", u.Variant())
			}
		})
	}
}

func TestNewV3AllNamespaces(t *testing.T) {
	namespaces := []struct {
		name string
		ns   UUID
	}{
		{"DNS", NamespaceDNS},
		{"URL", NamespaceURL},
		{"OID", NamespaceOID},
		{"X500", NamespaceX500},
	}
	for _, tt := range namespaces {
		t.Run(tt.name, func(t *testing.T) {
			u := NewV3(tt.ns, "test")
			if u.Version() != Version3 {
				t.Errorf("Version = %v, want V3", u.Version())
			}
			if u.Variant() != VariantRFC9562 {
				t.Errorf("Variant = %v, want RFC9562", u.Variant())
			}
		})
	}
}

func TestNewV8(t *testing.T) {
	var data [16]byte
	for i := range data {
		data[i] = byte(i)
	}
	u := NewV8(data)
	if u.Version() != Version8 {
		t.Errorf("NewV8().Version() = %v, want V8", u.Version())
	}
	if u.Variant() != VariantRFC9562 {
		t.Errorf("NewV8().Variant() = %v, want RFC9562", u.Variant())
	}
	// Check that non-version/variant bits are preserved
	if u[0] != 0x00 || u[1] != 0x01 || u[2] != 0x02 || u[3] != 0x03 {
		t.Errorf("unexpected first 4 bytes: %x", u[:4])
	}
}

func TestNewV8Deterministic(t *testing.T) {
	var data [16]byte
	data[0] = 0xab
	a := NewV8(data)
	b := NewV8(data)
	if a != b {
		t.Errorf("NewV8 should be deterministic")
	}
}

func TestNewV5DifferentFromV3(t *testing.T) {
	v3 := NewV3(NamespaceDNS, "example.com")
	v5 := NewV5(NamespaceDNS, "example.com")
	if v3 == v5 {
		t.Errorf("V3 and V5 should produce different UUIDs for same input")
	}
}
