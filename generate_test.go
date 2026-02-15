package uuid

import (
	"slices"
	"testing"
	"testing/cryptotest"
	"testing/synctest"
	"time"
)

func TestNewV4(t *testing.T) {
	cryptotest.SetGlobalRandom(t, 42)

	u := NewV4()
	if u.Version() != V4 {
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
	if u.Version() != V3 {
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
	if u.Version() != V3 {
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
	if u.Version() != V5 {
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
			if u.Version() != V5 {
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
			if u.Version() != V3 {
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
	if u.Version() != V8 {
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

func TestNewV7Version(t *testing.T) {
	u := NewV7()
	if u.Version() != V7 {
		t.Errorf("NewV7().Version() = %v, want V7", u.Version())
	}
	if u.Variant() != VariantRFC9562 {
		t.Errorf("NewV7().Variant() = %v, want RFC9562", u.Variant())
	}
}

func TestNewV7Uniqueness(t *testing.T) {
	seen := make(map[UUID]bool)
	for range 1000 {
		u := NewV7()
		if seen[u] {
			t.Fatalf("duplicate V7 UUID: %s", u)
		}
		seen[u] = true
	}
}

func TestNewV7Monotonic(t *testing.T) {
	gen := NewGenerator()
	prev := gen.NewV7()
	for range 100 {
		curr := gen.NewV7()
		if Compare(curr, prev) <= 0 {
			t.Fatalf("V7 not monotonic: %s <= %s", curr, prev)
		}
		prev = curr
	}
}

func TestNewV7MonotonicSameMillisecond(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cryptotest.SetGlobalRandom(t, 99)

		gen := NewGenerator()
		// Generate multiple UUIDs without advancing time — all in same millisecond
		a := gen.NewV7()
		b := gen.NewV7()
		c := gen.NewV7()

		if Compare(a, b) >= 0 {
			t.Errorf("expected a < b: %s >= %s", a, b)
		}
		if Compare(b, c) >= 0 {
			t.Errorf("expected b < c: %s >= %s", b, c)
		}

		// Millisecond timestamps are the same (sub-ms ordering is in rand_a)
		ta := a.Time()
		tb := b.Time()
		tc := c.Time()
		if !ta.Equal(tb) {
			t.Errorf("expected same ms timestamp: a=%v, b=%v", ta, tb)
		}
		if !tb.Equal(tc) {
			t.Errorf("expected same ms timestamp: b=%v, c=%v", tb, tc)
		}
	})
}

func TestNewV7TimestampAdvances(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		gen := NewGenerator()
		a := gen.NewV7()

		// Advance fake clock by 100ms
		time.Sleep(100 * time.Millisecond)

		b := gen.NewV7()
		if Compare(b, a) <= 0 {
			t.Errorf("V7 should be monotonic after time advance: %s <= %s", b, a)
		}

		diff := b.Time().Sub(a.Time())
		if diff < 100*time.Millisecond {
			t.Errorf("expected >= 100ms difference, got %v", diff)
		}
	})
}

func TestNewV7Sortable(t *testing.T) {
	gen := NewGenerator()
	uuids := make([]UUID, 100)
	for i := range uuids {
		uuids[i] = gen.NewV7()
	}

	sorted := slices.IsSortedFunc(uuids, Compare)
	if !sorted {
		t.Errorf("V7 UUIDs should be naturally sorted")
	}
}

func TestNewV7GeneratorIsolation(t *testing.T) {
	gen1 := NewGenerator()
	gen2 := NewGenerator()

	u1 := gen1.NewV7()
	u2 := gen2.NewV7()

	// Different generators should produce different UUIDs
	if u1 == u2 {
		t.Errorf("different generators produced same UUID: %s", u1)
	}
}

func TestNewV7Deterministic(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		cryptotest.SetGlobalRandom(t, 42)
		gen := NewGenerator()
		a := gen.NewV7()

		cryptotest.SetGlobalRandom(t, 42)
		gen2 := NewGenerator()
		b := gen2.NewV7()

		if a != b {
			t.Errorf("deterministic V7 failed: %s != %s", a, b)
		}
	})
}

func TestNewV7PackageLevelUsesDefault(t *testing.T) {
	u := NewV7()
	if u.Version() != V7 {
		t.Errorf("package-level NewV7().Version() = %v, want V7", u.Version())
	}
}

func TestNewV7ConcurrentSafety(t *testing.T) {
	gen := NewGenerator()
	const n = 100
	results := make(chan UUID, n)

	for range n {
		go func() {
			results <- gen.NewV7()
		}()
	}

	seen := make(map[UUID]bool, n)
	for range n {
		u := <-results
		if seen[u] {
			t.Fatalf("duplicate UUID from concurrent generation: %s", u)
		}
		seen[u] = true
	}
}
