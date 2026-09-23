package uuid

import (
	"crypto/rand"
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
		want := refV5(NamespaceDNS, name)
		if got := NewV5(NamespaceDNS, name); got != want {
			t.Errorf("NewV5(len %d) = %s, want %s", n, got, want)
		}
		if got := NewV5Bytes(NamespaceDNS, []byte(name)); got != want {
			t.Errorf("NewV5Bytes(len %d) = %s, want %s", n, got, want)
		}
	}
}

func TestNewV5ZeroAlloc(t *testing.T) {
	name := []byte("www.example.com")
	longName := strings.Repeat("x", 1000) // takes the streaming path
	for fn, run := range map[string]func(){
		"NewV5":           func() { _ = NewV5(NamespaceDNS, "www.example.com") },
		"NewV5Bytes":      func() { _ = NewV5Bytes(NamespaceDNS, name) },
		"NewV5 long name": func() { _ = NewV5(NamespaceDNS, longName) },
		// A caller's stack buffer must stay on the stack, which requires
		// that NewV5Bytes's name parameter does not escape.
		"NewV5Bytes stack buffer": func() {
			var buf [64]byte
			n := copy(buf[:], "www.example.com")
			_ = NewV5Bytes(NamespaceDNS, buf[:n])
		},
	} {
		if allocs := testing.AllocsPerRun(100, run); allocs != 0 {
			t.Errorf("%s allocs = %v, want 0", fn, allocs)
		}
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

// seqOf decodes the 60-bit V7 ordering value, ms<<12 | rand_a, that the
// monotonic counter increments.
func seqOf(u UUID) int64 {
	ms := int64(u[0])<<40 | int64(u[1])<<32 | int64(u[2])<<24 |
		int64(u[3])<<16 | int64(u[4])<<8 | int64(u[5])
	return ms<<12 | int64(u[6]&0x0f)<<8 | int64(u[7])
}

// wantSeq computes the ordering value for t straight from RFC 9562 Section
// 6.2 Method 3, independently of the generators' own arithmetic.
func wantSeq(t time.Time) int64 {
	return t.UnixMilli()<<12 | int64(t.Nanosecond()%1_000_000)*4096/1_000_000
}

// v7Source is one way to draw V7 UUIDs from a single monotonic state.
type v7Source struct {
	name string
	// open returns a draw function backed by fresh state, and a pointer to
	// that state's last reserved ordering value.
	open func() (draw func(n int) []UUID, last *int64)
}

func v7Sources() []v7Source {
	loop := func(next func() UUID) func(n int) []UUID {
		return func(n int) []UUID {
			ids := make([]UUID, n)
			for i := range ids {
				ids[i] = next()
			}
			return ids
		}
	}
	return []v7Source{
		{"Generator", func() (func(int) []UUID, *int64) {
			g := NewGenerator()
			return loop(g.NewV7), &g.lastSeq
		}},
		{"zero-value Generator", func() (func(int) []UUID, *int64) {
			var g Generator
			return loop(g.NewV7), &g.lastSeq
		}},
		{"Generator.NewV7Batch", func() (func(int) []UUID, *int64) {
			g := NewGenerator()
			return g.NewV7Batch, &g.lastSeq
		}},
		{"Generator.FillV7", func() (func(int) []UUID, *int64) {
			g := NewGenerator()
			return func(n int) []UUID {
				dst := make([]UUID, n)
				g.FillV7(dst)
				return dst
			}, &g.lastSeq
		}},
		{"NewPoolFor", func() (func(int) []UUID, *int64) {
			g := NewGenerator()
			p := NewPoolFor(g)
			return loop(p.NewV7), &g.lastSeq
		}},
		{"zero-value Pool", func() (func(int) []UUID, *int64) {
			// A zero-value Pool orders through the package-level default
			// generator. Start it fresh so values follow this test's clock;
			// tests do not run in parallel, so nothing else is using it.
			defaultGen.mu.Lock()
			defaultGen.lastSeq = 0
			defaultGen.mu.Unlock()
			var p Pool
			return loop(p.NewV7), &defaultGen.lastSeq
		}},
	}
}

// checkRun fails t unless ids are well-formed V7 UUIDs whose ordering values
// run consecutively from start, which is what every source must produce
// while the clock stands still.
func checkRun(t *testing.T, ids []UUID, start int64) {
	t.Helper()
	for i, u := range ids {
		if u.Version() != V7 || u.Variant() != VariantRFC9562 {
			t.Fatalf("ids[%d] = %s: version %v, variant %v", i, u, u.Version(), u.Variant())
		}
		if got, want := seqOf(u), start+int64(i); got != want {
			t.Fatalf("ids[%d] ordering value = %d, want %d (start + %d)", i, got, want, i)
		}
	}
}

func TestV7CounterCarriesIntoNextMillisecond(t *testing.T) {
	for _, src := range v7Sources() {
		t.Run(src.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				draw, _ := src.open()
				now := time.Now()
				start := wantSeq(now)

				// With the clock frozen, 5000 UUIDs exceed the 4096 values
				// one millisecond holds, so the counter must carry into the
				// millisecond field while staying strictly ordered.
				ids := draw(5000)
				checkRun(t, ids, start)

				carry := 4096 - int(start&0xFFF) // first index in the next millisecond
				before, _ := ids[carry-1].Time()
				after, _ := ids[carry].Time()
				if ms := now.UnixMilli(); before.UnixMilli() != ms || after.UnixMilli() != ms+1 {
					t.Errorf("carry at index %d: Time() = %d ms then %d ms, want %d then %d",
						carry, before.UnixMilli(), after.UnixMilli(), ms, ms+1)
				}

				// Once the wall clock passes the counter, generation follows
				// the clock again instead of continuing the counter.
				synctest.Sleep(10 * time.Millisecond)
				next := draw(1)[0]
				if got, want := seqOf(next), wantSeq(time.Now()); got != want {
					t.Errorf("after the clock caught up: ordering value = %d, want %d from the clock", got, want)
				}
				if Compare(next, ids[len(ids)-1]) <= 0 {
					t.Errorf("after the clock caught up: %s does not sort after %s", next, ids[len(ids)-1])
				}
			})
		})
	}
}

func TestV7ClockBehindLastValue(t *testing.T) {
	for _, src := range v7Sources() {
		t.Run(src.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				draw, last := src.open()
				now := time.Now()

				// A value reserved an hour ahead is what a clock stepped back
				// by an hour looks like: new UUIDs must continue from it.
				ahead := wantSeq(now.Add(time.Hour))
				*last = ahead
				checkRun(t, draw(3), ahead+1)

				// When the clock passes that point, it takes over again.
				synctest.Sleep(2 * time.Hour)
				if got, want := seqOf(draw(1)[0]), wantSeq(time.Now()); got != want {
					t.Errorf("after the clock caught up: ordering value = %d, want %d", got, want)
				}
			})
		})
	}
}

func TestV7ConcurrentDrawsNeverShareAValue(t *testing.T) {
	for _, src := range v7Sources() {
		t.Run(src.name, func(t *testing.T) {
			draw, _ := src.open()
			const goroutines, perDraw = 50, 20
			results := make(chan []UUID, goroutines)
			for range goroutines {
				go func() { results <- draw(perDraw) }()
			}

			// Concurrent callers are not ordered relative to each other, but
			// each draw is, and no two UUIDs from one source may share an
			// ordering value, which also makes them unique.
			seen := make(map[int64]bool, goroutines*perDraw)
			for range goroutines {
				ids := <-results
				if !slices.IsSortedFunc(ids, Compare) {
					t.Errorf("a single draw is not ordered: %v", ids)
				}
				for _, u := range ids {
					if seen[seqOf(u)] {
						t.Fatalf("ordering value of %s handed out twice", u)
					}
					seen[seqOf(u)] = true
				}
			}
		})
	}
}

func TestV7BatchAndSingleShareOneCounter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		gen := NewGenerator()
		start := wantSeq(time.Now())
		ids := slices.Concat(
			[]UUID{gen.NewV7()},
			gen.NewV7Batch(5),
			[]UUID{gen.NewV7()},
			gen.NewV7Batch(3),
		)
		checkRun(t, ids, start)
	})
}

func TestFillV4(t *testing.T) {
	dst := make([]UUID, 1000)
	FillV4(dst)
	seen := make(map[UUID]bool, len(dst))
	for i, u := range dst {
		if u.Version() != V4 || u.Variant() != VariantRFC9562 {
			t.Fatalf("dst[%d] = %s: version %v, variant %v", i, u, u.Version(), u.Variant())
		}
		if seen[u] {
			t.Fatalf("dst[%d] = %s repeats an earlier UUID", i, u)
		}
		seen[u] = true
	}

	// FillV4 and NewV4Batch read the same random stream the same way.
	cryptotest.SetGlobalRandom(t, 7)
	want := NewV4Batch(10)
	cryptotest.SetGlobalRandom(t, 7)
	got := make([]UUID, 10)
	FillV4(got)
	if !slices.Equal(got, want) {
		t.Errorf("FillV4 = %v, want NewV4Batch's %v", got, want)
	}
}

func TestFillEmpty(t *testing.T) {
	FillV4(nil)
	FillV4([]UUID{})

	gen := NewGenerator()
	gen.FillV7(nil)
	gen.FillV7([]UUID{})
	if gen.lastSeq != 0 {
		t.Errorf("FillV7 of nothing advanced the generator to %d", gen.lastSeq)
	}
}

func TestFillV7TakesRandBFromTheStream(t *testing.T) {
	// FillV7 reads rand_b into the upper half of dst and encodes in place.
	// With a deterministic random stream, every UUID's rand_b must be its
	// own 8-byte slice of that stream: an overlap mistake would hand a UUID
	// bytes that were already overwritten. Small and odd sizes exercise the
	// boundary where the two halves meet.
	for n := 1; n <= 300; n++ {
		cryptotest.SetGlobalRandom(t, uint64(n))
		stream := make([]byte, n*8)
		_, _ = rand.Read(stream)

		cryptotest.SetGlobalRandom(t, uint64(n))
		dst := make([]UUID, n)
		NewGenerator().FillV7(dst)

		for i, u := range dst {
			want := [8]byte(stream[i*8:])
			want[0] = want[0]&0x3f | 0x80 // FillV7 sets the variant bits
			if got := [8]byte(u[8:]); got != want {
				t.Fatalf("n=%d: dst[%d] rand_b = %x, want %x from the stream", n, i, got, want)
			}
		}
	}
}

func TestFillV7PackageLevel(t *testing.T) {
	// The package-level FillV7 shares the default generator with NewV7.
	before := NewV7()
	dst := make([]UUID, 5)
	FillV7(dst)
	after := NewV7()
	all := slices.Concat([]UUID{before}, dst, []UUID{after})
	if !slices.IsSortedFunc(all, Compare) {
		t.Errorf("NewV7, FillV7, NewV7 are not ordered: %v", all)
	}
}

func TestPoolUsesItsGenerator(t *testing.T) {
	t.Cleanup(func() {
		defaultGen.mu.Lock()
		defaultGen.lastSeq = 0
		defaultGen.mu.Unlock()
	})
	g := NewGenerator()
	for name, tc := range map[string]struct {
		pool *Pool
		gen  *Generator
	}{
		"NewPool":         {NewPool(), defaultGen},
		"zero value":      {&Pool{}, defaultGen},
		"NewPoolFor(nil)": {NewPoolFor(nil), defaultGen},
		"NewPoolFor(g)":   {NewPoolFor(g), g},
	} {
		// With the generator an hour ahead of the clock, the pool's next
		// UUID must continue from it; a pool with its own state would
		// follow the clock instead.
		ahead := wantSeq(time.Now().Add(time.Hour))
		tc.gen.mu.Lock()
		tc.gen.lastSeq = ahead
		tc.gen.mu.Unlock()
		if got := seqOf(tc.pool.NewV7()); got != ahead+1 {
			t.Errorf("%s: ordering value = %d, want %d, right after its generator's", name, got, ahead+1)
		}
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
	dst := make([]UUID, 1000)
	for name, fn := range map[string]func(){
		"FillV4":           func() { FillV4(dst) },
		"FillV7":           func() { FillV7(dst) },
		"Generator.FillV7": func() { gen.FillV7(dst) },
		"NewV4":            func() { _ = NewV4() },
		"NewV7":            func() { _ = NewV7() },
		"NewV7At":          func() { _ = NewV7At(at) },
		"Generator":        func() { _ = gen.NewV7() },
		"Pool.NewV4":       func() { _ = pool.NewV4() },
		"Pool.NewV7":       func() { _ = pool.NewV7() },
	} {
		t.Run(name, func(t *testing.T) {
			if allocs := testing.AllocsPerRun(100, fn); allocs != 0 {
				t.Errorf("%s allocs = %v, want 0", name, allocs)
			}
		})
	}
}
