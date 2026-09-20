package uuid

import (
	"crypto/sha1"
	"slices"
	"strings"
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

func TestNewV4Batch(t *testing.T) {
	uuids := NewV4Batch(100)
	if len(uuids) != 100 {
		t.Fatalf("NewV4Batch(100) returned %d UUIDs", len(uuids))
	}
	seen := make(map[UUID]bool, 100)
	for i, u := range uuids {
		if u.Version() != V4 {
			t.Errorf("uuids[%d].Version() = %v, want V4", i, u.Version())
		}
		if u.Variant() != VariantRFC9562 {
			t.Errorf("uuids[%d].Variant() = %v, want RFC9562", i, u.Variant())
		}
		if u.IsNil() {
			t.Errorf("uuids[%d] should not be nil", i)
		}
		if seen[u] {
			t.Fatalf("duplicate UUID in batch at index %d: %s", i, u)
		}
		seen[u] = true
	}
}

func TestNewV4BatchZero(t *testing.T) {
	uuids := NewV4Batch(0)
	if len(uuids) != 0 {
		t.Fatalf("NewV4Batch(0) returned %d UUIDs, want 0", len(uuids))
	}
}

func TestNewV4BatchNegative(t *testing.T) {
	uuids := NewV4Batch(-1)
	if uuids != nil {
		t.Fatalf("NewV4Batch(-1) = %v, want nil", uuids)
	}
}

func TestNewV4BatchDeterministic(t *testing.T) {
	cryptotest.SetGlobalRandom(t, 77)
	a := NewV4Batch(10)

	cryptotest.SetGlobalRandom(t, 77)
	b := NewV4Batch(10)

	for i := range a {
		if a[i] != b[i] {
			t.Errorf("batch[%d] mismatch with same seed: %s != %s", i, a[i], b[i])
		}
	}
}

func TestPoolNewV4(t *testing.T) {
	pool := NewPool()
	seen := make(map[UUID]bool, 1000)
	for range 1000 {
		u := pool.NewV4()
		if u.Version() != V4 {
			t.Errorf("Pool.NewV4().Version() = %v, want V4", u.Version())
		}
		if u.Variant() != VariantRFC9562 {
			t.Errorf("Pool.NewV4().Variant() = %v, want RFC9562", u.Variant())
		}
		if seen[u] {
			t.Fatalf("duplicate UUID from pool: %s", u)
		}
		seen[u] = true
	}
}

func TestPoolZeroValue(t *testing.T) {
	var p Pool
	for i := range 3 {
		u := p.NewV4()
		if u.IsNil() {
			t.Fatalf("zero-value Pool.NewV4() #%d returned Nil", i)
		}
		if u.Version() != V4 || u.Variant() != VariantRFC9562 {
			t.Fatalf("zero-value Pool.NewV4() #%d = %s, bad version/variant", i, u)
		}
	}
	for i := range 3 {
		u := p.NewV7()
		if u.Version() != V7 || u.Variant() != VariantRFC9562 {
			t.Fatalf("zero-value Pool.NewV7() #%d = %s, bad version/variant", i, u)
		}
		var zero [7]byte
		if [7]byte(u[9:16]) == zero {
			t.Fatalf("zero-value Pool.NewV7() #%d has all-zero rand_b: %s", i, u)
		}
	}
}

func TestPoolNewV4ConcurrentSafety(t *testing.T) {
	pool := NewPool()
	const n = 500
	results := make(chan UUID, n)

	for range n {
		go func() {
			results <- pool.NewV4()
		}()
	}

	seen := make(map[UUID]bool, n)
	for range n {
		u := <-results
		if seen[u] {
			t.Fatalf("duplicate UUID from concurrent pool: %s", u)
		}
		seen[u] = true
	}
}

func TestPoolNewV7(t *testing.T) {
	pool := NewPool()
	seen := make(map[UUID]bool, 1000)
	for range 1000 {
		u := pool.NewV7()
		if u.Version() != V7 {
			t.Errorf("Pool.NewV7().Version() = %v, want V7", u.Version())
		}
		if u.Variant() != VariantRFC9562 {
			t.Errorf("Pool.NewV7().Variant() = %v, want RFC9562", u.Variant())
		}
		if seen[u] {
			t.Fatalf("duplicate UUID from pool V7: %s", u)
		}
		seen[u] = true
	}
}

func TestPoolNewV7Monotonic(t *testing.T) {
	pool := NewPool()
	prev := pool.NewV7()
	for range 100 {
		curr := pool.NewV7()
		if Compare(curr, prev) <= 0 {
			t.Fatalf("Pool V7 not monotonic: %s <= %s", curr, prev)
		}
		prev = curr
	}
}

func TestPoolNewV7ConcurrentSafety(t *testing.T) {
	pool := NewPool()
	const n = 500
	results := make(chan UUID, n)

	for range n {
		go func() {
			results <- pool.NewV7()
		}()
	}

	seen := make(map[UUID]bool, n)
	for range n {
		u := <-results
		if seen[u] {
			t.Fatalf("duplicate UUID from concurrent pool V7: %s", u)
		}
		seen[u] = true
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

// refV5 computes a V5 UUID straight from the RFC 9562 definition
// (SHA-1 of namespace||name) using the streaming hash API.
func refV5(ns UUID, name string) UUID {
	h := sha1.New()
	h.Write(ns[:])
	h.Write([]byte(name))
	var u UUID
	copy(u[:], h.Sum(nil))
	u[6] = (u[6] & 0x0f) | 0x50
	u[8] = (u[8] & 0x3f) | 0x80
	return u
}

func TestNewV5NameLengths(t *testing.T) {
	// Names up to v5StackBuf-16 bytes use the stack buffer; longer names
	// take the streaming path. Both must agree with the RFC definition.
	for _, n := range []int{0, 1, 15, v5StackBuf - 17, v5StackBuf - 16, v5StackBuf - 15, 1000} {
		name := strings.Repeat("x", n)
		if got, want := NewV5(NamespaceDNS, name), refV5(NamespaceDNS, name); got != want {
			t.Errorf("NewV5(len %d) = %s, want %s", n, got, want)
		}
	}
}

func TestNewV5ZeroAlloc(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		_ = NewV5(NamespaceDNS, "www.example.com")
	})
	if allocs != 0 {
		t.Errorf("NewV5 allocs = %v, want 0", allocs)
	}
}

func TestNewV5CustomNamespace(t *testing.T) {
	ns := MustParse("12345678-1234-1234-1234-123456789abc")
	u := NewV5(ns, "hello")
	if u.Version() != V5 {
		t.Errorf("Version = %v, want V5", u.Version())
	}
	if u.Variant() != VariantRFC9562 {
		t.Errorf("Variant = %v, want RFC9562", u.Variant())
	}
	// Same input must produce same output
	u2 := NewV5(ns, "hello")
	if u != u2 {
		t.Errorf("determinism failed")
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
		ta, _ := a.Time()
		tb, _ := b.Time()
		tc, _ := c.Time()
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
		synctest.Sleep(100 * time.Millisecond)

		b := gen.NewV7()
		if Compare(b, a) <= 0 {
			t.Errorf("V7 should be monotonic after time advance: %s <= %s", b, a)
		}

		ta, _ := a.Time()
		tb, _ := b.Time()
		diff := tb.Sub(ta)
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

func TestNewV7BatchPackageLevel(t *testing.T) {
	uuids := NewV7Batch(10)
	if len(uuids) != 10 {
		t.Fatalf("NewV7Batch(10) returned %d UUIDs", len(uuids))
	}
	if !slices.IsSortedFunc(uuids, Compare) {
		t.Errorf("package-level NewV7Batch should be monotonically increasing")
	}
	// Shares the default generator with NewV7: a subsequent single UUID
	// must sort after the batch.
	single := NewV7()
	if Compare(single, uuids[len(uuids)-1]) <= 0 {
		t.Errorf("NewV7() after batch should sort after it: %s <= %s", single, uuids[len(uuids)-1])
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

func TestNewV7Batch(t *testing.T) {
	gen := NewGenerator()
	uuids := gen.NewV7Batch(100)
	if len(uuids) != 100 {
		t.Fatalf("NewV7Batch(100) returned %d UUIDs", len(uuids))
	}
	seen := make(map[UUID]bool, 100)
	for i, u := range uuids {
		if u.Version() != V7 {
			t.Errorf("uuids[%d].Version() = %v, want V7", i, u.Version())
		}
		if u.Variant() != VariantRFC9562 {
			t.Errorf("uuids[%d].Variant() = %v, want RFC9562", i, u.Variant())
		}
		if seen[u] {
			t.Fatalf("duplicate UUID in V7 batch at index %d: %s", i, u)
		}
		seen[u] = true
	}
}

func TestNewV7BatchMonotonic(t *testing.T) {
	gen := NewGenerator()
	uuids := gen.NewV7Batch(100)

	if !slices.IsSortedFunc(uuids, Compare) {
		t.Errorf("V7 batch UUIDs should be monotonically increasing")
	}
}

func TestNewV7BatchZero(t *testing.T) {
	gen := NewGenerator()
	uuids := gen.NewV7Batch(0)
	if len(uuids) != 0 {
		t.Fatalf("NewV7Batch(0) returned %d UUIDs, want 0", len(uuids))
	}
}

func TestNewV7BatchNegative(t *testing.T) {
	gen := NewGenerator()
	uuids := gen.NewV7Batch(-1)
	if uuids != nil {
		t.Fatalf("NewV7Batch(-1) = %v, want nil", uuids)
	}
}

func TestNewV7BatchMonotonicAcrossCalls(t *testing.T) {
	gen := NewGenerator()
	batch1 := gen.NewV7Batch(10)
	batch2 := gen.NewV7Batch(10)

	lastOfBatch1 := batch1[len(batch1)-1]
	firstOfBatch2 := batch2[0]
	if Compare(firstOfBatch2, lastOfBatch1) <= 0 {
		t.Errorf("batch2[0] should be > batch1[9]: %s <= %s", firstOfBatch2, lastOfBatch1)
	}
}

func TestNewV7BatchMonotonicSameMillisecond(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		gen := NewGenerator()
		// First call advances lastSeq
		batch1 := gen.NewV7Batch(5)
		// Second call at the same fake-clock time must hit the seq <= lastSeq fallback
		batch2 := gen.NewV7Batch(5)

		all := slices.Concat(batch1, batch2)
		if !slices.IsSortedFunc(all, Compare) {
			t.Errorf("V7 batches at same clock time should be monotonically increasing")
		}
	})
}

func TestPoolNewV7MonotonicSameMillisecond(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		pool := NewPool()
		// Generate multiple UUIDs at the same fake-clock time
		// to force the seq <= p.v7seq fallback branch
		a := pool.NewV7()
		b := pool.NewV7()
		c := pool.NewV7()

		if Compare(a, b) >= 0 {
			t.Errorf("expected a < b: %s >= %s", a, b)
		}
		if Compare(b, c) >= 0 {
			t.Errorf("expected b < c: %s >= %s", b, c)
		}
	})
}

func TestNewV7BatchInterleavedWithSingle(t *testing.T) {
	gen := NewGenerator()
	batch := gen.NewV7Batch(10)
	single := gen.NewV7()

	lastOfBatch := batch[len(batch)-1]
	if Compare(single, lastOfBatch) <= 0 {
		t.Errorf("single NewV7 should be > last batch UUID: %s <= %s", single, lastOfBatch)
	}
}

func TestNewV7At(t *testing.T) {
	at := time.Date(2020, time.March, 14, 15, 9, 26, 535_897_932, time.UTC)
	u := NewV7At(at)

	if u.Version() != V7 {
		t.Errorf("Version() = %v, want V7", u.Version())
	}
	if u.Variant() != VariantRFC9562 {
		t.Errorf("Variant() = %v, want RFC9562", u.Variant())
	}
	got, ok := u.Time()
	if !ok {
		t.Fatal("Time() ok = false")
	}
	if want := at.Truncate(time.Millisecond); !got.Equal(want) {
		t.Errorf("Time() = %v, want %v", got, want)
	}

	// Sub-millisecond fraction per RFC 9562 Method 3: 897_932 ns * 4096 / 1e6.
	wantFrac := int64(897_932) * 4096 / 1_000_000
	gotFrac := int64(u[6]&0x0f)<<8 | int64(u[7])
	if gotFrac != wantFrac {
		t.Errorf("rand_a fraction = %d, want %d", gotFrac, wantFrac)
	}
}

func TestNewV7AtSortsAmongLive(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		gen := NewGenerator()
		now := time.Now()

		past := NewV7At(now.Add(-time.Hour))
		live := gen.NewV7()
		future := NewV7At(now.Add(time.Hour))

		if Compare(past, live) >= 0 {
			t.Errorf("past NewV7At should sort before live: %s >= %s", past, live)
		}
		if Compare(live, future) >= 0 {
			t.Errorf("live should sort before future NewV7At: %s >= %s", live, future)
		}
	})
}

func TestNewV7AtDoesNotAdvanceGenerator(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		gen := NewGenerator()
		now := time.Now()

		// Backfilling with a future timestamp must not push live UUIDs ahead.
		_ = NewV7At(now.Add(24 * time.Hour))
		live := gen.NewV7()

		got, _ := live.Time()
		if !got.Equal(now.Truncate(time.Millisecond)) {
			t.Errorf("live Time() = %v, want %v", got, now.Truncate(time.Millisecond))
		}
	})
}

func TestNewV7AtRandomTail(t *testing.T) {
	cryptotest.SetGlobalRandom(t, 42)
	at := time.Unix(1_700_000_000, 0)
	a := NewV7At(at)
	b := NewV7At(at)

	if [8]byte(a[:8]) != [8]byte(b[:8]) {
		t.Errorf("same t should give identical first 8 bytes: %x vs %x", a[:8], b[:8])
	}
	if a == b {
		t.Error("rand_b should differ between calls")
	}
}

func TestNewV7AtOutOfRangePanics(t *testing.T) {
	for _, tc := range []struct {
		name string
		at   time.Time
	}{
		{"zero time", time.Time{}},
		{"before epoch", time.Unix(-1, 0)},
		{"beyond 48 bits", time.UnixMilli(1 << 48)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("NewV7At(%v) did not panic", tc.at)
				}
			}()
			NewV7At(tc.at)
		})
	}

	// Boundary: the largest representable millisecond is valid.
	u := NewV7At(time.UnixMilli(1<<48 - 1))
	got, _ := u.Time()
	if got.UnixMilli() != 1<<48-1 {
		t.Errorf("Time().UnixMilli() = %d, want %d", got.UnixMilli(), int64(1<<48-1))
	}
}

// TestRandZeroAlloc enforces the zero-alloc guarantee for every generator
// that reads crypto/rand. Skipped under the race detector: see raceEnabled.
func TestRandZeroAlloc(t *testing.T) {
	if raceEnabled {
		t.Skip("crypto/rand.Read allocates under the race detector on Linux")
	}
	gen := NewGenerator()
	pool := NewPool()
	at := time.Now()
	for name, fn := range map[string]func(){
		"NewV4":      func() { _ = NewV4() },
		"NewV7":      func() { _ = NewV7() },
		"NewV7At":    func() { _ = NewV7At(at) },
		"Generator":  func() { _ = gen.NewV7() },
		"Pool.NewV4": func() { _ = pool.NewV4() },
		"Pool.NewV7": func() { _ = pool.NewV7() },
	} {
		t.Run(name, func(t *testing.T) {
			if allocs := testing.AllocsPerRun(100, fn); allocs != 0 {
				t.Errorf("%s allocs = %v, want 0", name, allocs)
			}
		})
	}
}
