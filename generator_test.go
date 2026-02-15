package uuid

import (
	"slices"
	"testing"
	"testing/cryptotest"
	"testing/synctest"
	"time"
)

func TestNewV7Version(t *testing.T) {
	u := NewV7()
	if u.Version() != Version7 {
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

		// Timestamps should be strictly increasing
		ta := a.Time()
		tb := b.Time()
		tc := c.Time()
		if !tb.After(ta) {
			t.Errorf("expected b.Time > a.Time: %v <= %v", tb, ta)
		}
		if !tc.After(tb) {
			t.Errorf("expected c.Time > b.Time: %v <= %v", tc, tb)
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
	if u.Version() != Version7 {
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
