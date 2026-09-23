package uuid

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"sync"
	"time"
	"unsafe"
)

// NewV4 returns a new random (Version 4) UUID.
// It reads from crypto/rand, which cannot fail since Go 1.24.
func NewV4() UUID {
	var u UUID
	_, _ = rand.Read(u[:])
	stamp(&u, V4)
	return u
}

// v5StackBuf is the size of the stack buffer used to hash namespace||name
// in NewV5 with a single sha1.Sum call. Longer names are streamed into the
// hash through a buffer of the same size.
const v5StackBuf = 256

// NewV5 returns a deterministic Version 5 (SHA-1) UUID for the given namespace and name.
// It allocates nothing, whatever the length of name.
func NewV5(namespace UUID, name string) UUID {
	var sum [sha1.Size]byte
	if len(name) <= v5StackBuf-len(namespace) {
		// Hash namespace||name from a stack buffer in one call.
		var buf [v5StackBuf]byte
		copy(buf[:], namespace[:])
		n := copy(buf[len(namespace):], name)
		sum = sha1.Sum(buf[:len(namespace)+n])
	} else {
		// Feed the name through a local buffer rather than passing it to
		// the hash's Write, an interface call that would make name escape
		// and move callers' stack buffers (see NewV5Bytes) to the heap.
		h := sha1.New()
		h.Write(namespace[:])
		var chunk [v5StackBuf]byte
		for rest := name; rest != ""; {
			n := copy(chunk[:], rest)
			h.Write(chunk[:n])
			rest = rest[n:]
		}
		h.Sum(sum[:0])
	}

	var u UUID
	copy(u[:], sum[:16])
	stamp(&u, V5)
	return u
}

// NewV5Bytes is [NewV5] for a name held in a byte slice, so callers need not
// convert it to a string first. It returns the same UUID as NewV5 for the
// same bytes and allocates nothing, even when name is a caller's stack buffer.
func NewV5Bytes(namespace UUID, name []byte) UUID {
	if len(name) == 0 {
		return NewV5(namespace, "")
	}
	// NewV5 only hashes name and never retains it, so viewing the bytes as a
	// string without copying is safe for the duration of the call.
	return NewV5(namespace, unsafe.String(&name[0], len(name)))
}

// NewV4Batch returns n random (Version 4) UUIDs.
// It amortizes the cost of crypto/rand by reading all random bytes in a
// single call, making it significantly faster than calling [NewV4] in a loop.
// It returns nil if n <= 0. To reuse a buffer instead, see [FillV4].
func NewV4Batch(n int) []UUID {
	if n <= 0 {
		return nil
	}
	uuids := make([]UUID, n)
	FillV4(uuids)
	return uuids
}

// FillV4 overwrites every element of dst with a new random (Version 4) UUID.
// It is the allocation-free form of [NewV4Batch] for callers that reuse a
// buffer, and like it reads all random bytes in a single crypto/rand call.
func FillV4(dst []UUID) {
	if len(dst) == 0 {
		return
	}
	_, _ = rand.Read(rawBytes(dst))
	for i := range dst {
		stamp(&dst[i], V4)
	}
}

// rawBytes returns the contiguous backing memory of a non-empty UUID slice
// as a byte slice, so crypto/rand can fill it directly without an
// intermediate buffer. A []UUID is a contiguous array of [16]byte, so the
// resulting slice covers exactly len(uuids)*16 bytes of valid memory.
func rawBytes(uuids []UUID) []byte {
	return unsafe.Slice(&uuids[0][0], len(uuids)*16)
}

// Pool amortizes the cost of crypto/rand by pre-generating random bytes
// in bulk. It provides high-throughput [Pool.NewV4] and [Pool.NewV7] methods
// that are functionally equivalent to the package-level functions.
// Multiple goroutines may safely call methods concurrently.
//
// A Pool only buffers randomness; the ordering of its V7 UUIDs comes from a
// [Generator]. [NewPool] and the zero value use the package-level default
// generator, so Pool.NewV7 and [NewV7] produce one ordered sequence, and so
// do separate Pools. Use [NewPoolFor] to order a Pool's V7 UUIDs with a
// specific Generator instead.
//
// Because Pool buffers pre-generated randomness in process memory, it is
// not fork-safe: a forked process or a cloned/restored VM snapshot can
// duplicate the buffer, causing both copies to emit identical UUIDs.
// Use the package-level functions where fork or VM-clone safety matters.
//
// The zero value is ready to use; [NewPool] is equivalent to &Pool{}.
type Pool struct {
	mu sync.Mutex

	// V4: fully pre-stamped UUIDs ready to hand out.
	v4buf  [poolSize]UUID
	v4left int // UUIDs remaining in v4buf; zero triggers a refill

	// V7: pre-generated random bytes for rand_b (bytes 8–15).
	// Timestamp + monotonic sequence are computed live per call.
	v7rand [poolSize * 8]byte
	v7left int // 8-byte chunks remaining in v7rand; zero triggers a refill

	gen *Generator // orders V7 UUIDs; nil means the package-level default
}

const poolSize = 256

// NewPool returns a new [Pool] that amortizes crypto/rand overhead. Its V7
// UUIDs are ordered with [NewV7] through the package-level default
// generator; it is equivalent to &Pool{}.
func NewPool() *Pool {
	return &Pool{}
}

// NewPoolFor returns a new [Pool] whose V7 UUIDs continue gen's sequence,
// ordered with gen's own UUIDs and those of every other Pool created for
// gen. A nil gen means the package-level default generator, as for
// [NewPool].
func NewPoolFor(gen *Generator) *Pool {
	return &Pool{gen: gen}
}

func (p *Pool) refillV4() {
	_, _ = rand.Read(rawBytes(p.v4buf[:]))
	for i := range p.v4buf {
		stamp(&p.v4buf[i], V4)
	}
	p.v4left = poolSize
}

func (p *Pool) refillV7() {
	_, _ = rand.Read(p.v7rand[:])
	p.v7left = poolSize
}

// NewV4 returns a new random (Version 4) UUID from the pool.
// It is functionally equivalent to the package-level [NewV4] but
// amortizes the crypto/rand overhead across pool refills.
func (p *Pool) NewV4() UUID {
	p.mu.Lock()
	if p.v4left == 0 {
		p.refillV4()
	}
	u := p.v4buf[poolSize-p.v4left]
	p.v4left--
	p.mu.Unlock()
	return u
}

// NewV7 returns a new Version 7 UUID from the pool.
// It is functionally equivalent to [Generator.NewV7] on the pool's
// generator, and ordered with it, but amortizes the crypto/rand overhead by
// buffering random bytes for the rand_b field. Timestamps are computed live
// to remain accurate, though under sustained bursts they may run slightly
// ahead of the wall clock (see [Generator.NewV7]).
func (p *Pool) NewV7() UUID {
	// Read the clock before taking any lock, as Generator.NewV7 does, so
	// concurrent callers do not serialize on time.Now.
	seq := v7Seq(time.Now().UnixNano())

	var u UUID
	p.mu.Lock()
	if p.v7left == 0 {
		p.refillV7()
	}
	off := (poolSize - p.v7left) * 8
	copy(u[8:], p.v7rand[off:off+8])
	p.v7left--
	p.mu.Unlock()

	// The sequence comes from the generator, under its own lock, so this
	// UUID is ordered with everything else drawn from that generator.
	g := p.gen
	if g == nil {
		g = defaultGen
	}
	g.mu.Lock()
	seq = reserve(&g.lastSeq, seq, 1)
	g.mu.Unlock()

	setV7(&u, seq)
	return u
}

// NewV8 returns a Version 8 UUID constructed from user-provided data.
// The version and variant bits are set; all other 122 bits come from data.
// Uniqueness is the caller's responsibility per RFC 9562 Section 5.8.
func NewV8(data [16]byte) UUID {
	u := UUID(data)
	stamp(&u, V8)
	return u
}

// maxV7Millis is the exclusive upper bound of the 48-bit V7 timestamp field.
const maxV7Millis = 1 << 48

// NewV7At returns a Version 7 UUID whose timestamp fields encode t instead
// of the current time. It is intended for backfilling records that already
// have a creation time, so their keys sort among live V7 UUIDs at the
// right position.
//
// The 48-bit millisecond field and the 12-bit sub-millisecond fraction are
// derived from t exactly as [Generator.NewV7] derives them from the clock,
// so a UUID created "at" an instant sorts where a live UUID created at that
// instant would. The remaining 62 bits are random from crypto/rand. Two
// calls with the same t tie on their first 8 bytes and are ordered only by
// that random tail.
//
// NewV7At is stateless: it neither reads nor advances the monotonic
// state of any [Generator] or [Pool], so backfilling never pushes live
// UUIDs ahead of the wall clock.
//
// t must be representable in the 48-bit field, that is between the Unix
// epoch and roughly the year 10889; NewV7At panics otherwise. A zero
// [time.Time] is out of range, which turns an uninitialized field into an
// immediate panic rather than a silently wrong timestamp.
func NewV7At(t time.Time) UUID {
	ms := t.UnixMilli()
	if ms < 0 || ms >= maxV7Millis {
		panic("uuid: NewV7At: time out of range for a 48-bit millisecond timestamp")
	}
	// RFC 9562 Section 6.2 Method 3: sub-millisecond precision scaled to 12 bits.
	frac := int64(t.Nanosecond()%nanoPerMilli) * 4096 / nanoPerMilli

	var u UUID
	_, _ = rand.Read(u[8:])
	setV7(&u, ms<<12|frac)
	return u
}

// defaultGen is the package-level V7 generator, analogous to http.DefaultClient.
var defaultGen = NewGenerator()

// NewV7 returns a new Version 7 (Unix timestamp + random) UUID using the
// package-level default generator. For isolated monotonicity guarantees,
// create a dedicated [Generator] with [NewGenerator].
func NewV7() UUID {
	return defaultGen.NewV7()
}

// NewV7Batch returns n monotonically increasing Version 7 UUIDs using the
// package-level default generator. See [Generator.NewV7Batch].
func NewV7Batch(n int) []UUID {
	return defaultGen.NewV7Batch(n)
}

// FillV7 overwrites every element of dst with monotonically increasing
// Version 7 UUIDs from the package-level default generator. See
// [Generator.FillV7].
func FillV7(dst []UUID) {
	defaultGen.FillV7(dst)
}

// Generator produces Version 7 UUIDs with per-instance monotonicity.
// Multiple goroutines may safely call NewV7 concurrently on the same Generator.
//
// The zero value is ready to use; [NewGenerator] is equivalent to &Generator{}.
type Generator struct {
	mu      sync.Mutex
	lastSeq int64 // ms<<12 | seq for monotonicity
}

// NewGenerator returns a new V7 UUID generator with its own monotonicity state.
func NewGenerator() *Generator {
	return &Generator{}
}

const nanoPerMilli = 1_000_000

// v7Seq returns the V7 ordering value of a Unix time in nanoseconds: the
// millisecond timestamp shifted left 12 bits, plus the sub-millisecond
// fraction scaled to 12 bits (RFC 9562 Section 6.2 Method 3). Generators
// increment this value, so a carry out of the fraction advances the
// millisecond. Callers read the clock themselves, which keeps v7Seq cheap
// enough to inline.
func v7Seq(nano int64) int64 {
	return (nano/nanoPerMilli)<<12 | (nano%nanoPerMilli)*4096/nanoPerMilli
}

// reserve returns the first of n consecutive ordering values that start at
// seq, or right after *last if seq does not exceed it, and records the
// final value in *last. This keeps each generator strictly increasing even
// when the clock stalls or steps back. The caller holds the lock that
// guards *last.
func reserve(last *int64, seq int64, n int) int64 {
	if seq <= *last {
		seq = *last + 1
	}
	*last = seq + int64(n-1)
	return seq
}

// setV7 writes the V7 ordering value seq into bytes 0–7 of u: the 48-bit
// millisecond timestamp (seq>>12) big-endian in bytes 0–5, then the version
// and the 12-bit fraction (seq&0xFFF) in bytes 6–7. It also sets the variant
// bits; the random rand_b in bytes 8–15 is otherwise kept.
func setV7(u *UUID, seq int64) {
	binary.BigEndian.PutUint64(u[:8], uint64(seq>>12)<<16|uint64(seq&0xFFF))
	stamp(u, V7)
}

// stamp sets the version field (bits 48–51) to v and the variant field
// (bits 64–65) to RFC 9562, keeping the other 122 bits of u.
func stamp(u *UUID, v Version) {
	u[6] = (u[6] & 0x0f) | byte(v)<<4
	u[8] = (u[8] & 0x3f) | 0x80
}

// NewV7 returns a new Version 7 UUID.
//
// The UUID encodes a 48-bit Unix millisecond timestamp in bits 0–47 and
// 12 bits of sub-millisecond precision in the rand_a field (bits 48–59),
// computed per RFC 9562 Section 6.2 Method 3. The rand_b field (bytes 8–15,
// bits 64–127) is filled with random data from crypto/rand.
//
// When multiple UUIDs are generated faster than the clock resolution,
// the combined timestamp+seq counter is incremented to guarantee
// monotonicity within this Generator. Counter increments can carry into
// the millisecond field, so under sustained bursts the encoded timestamp
// (and thus [UUID.Time]) may run slightly ahead of the wall clock.
func (g *Generator) NewV7() UUID {
	var u UUID
	_, _ = rand.Read(u[8:])
	seq := v7Seq(time.Now().UnixNano())

	g.mu.Lock()
	seq = reserve(&g.lastSeq, seq, 1)
	g.mu.Unlock()

	setV7(&u, seq)
	return u
}

// NewV7Batch returns n Version 7 UUIDs that are monotonically increasing.
// It amortizes the cost of crypto/rand and [time.Now] across the batch,
// making it significantly faster than calling [Generator.NewV7] in a loop.
// It returns nil if n <= 0. To reuse a buffer instead, see
// [Generator.FillV7].
//
// All n UUIDs derive from one clock reading: consecutive counter values can
// carry into the millisecond field, so for large n the encoded timestamps
// may run slightly ahead of the wall clock.
func (g *Generator) NewV7Batch(n int) []UUID {
	if n <= 0 {
		return nil
	}
	uuids := make([]UUID, n)
	g.FillV7(uuids)
	return uuids
}

// FillV7 overwrites every element of dst with Version 7 UUIDs that are
// monotonically increasing, continuing this Generator's sequence. It is the
// allocation-free form of [Generator.NewV7Batch] for callers that reuse a
// buffer, with the same clock behavior: one clock reading for all of dst.
func (g *Generator) FillV7(dst []UUID) {
	if len(dst) == 0 {
		return
	}
	n := len(dst)

	// Read the 8 random rand_b bytes each UUID needs in one crypto/rand call,
	// into the upper half of dst's own memory. Encoding then runs front to
	// back: UUID i reads its bytes at src[8i:8i+8] before it writes
	// raw[16i:16i+16], and that write only reaches source bytes of UUIDs at
	// or before i (source offset 8n+8j lies in UUID i's slot only for
	// j = 2i-n or 2i-n+1, both at most i), which are already consumed. This
	// measured faster than a separate buffer, than chunking through a stack
	// buffer (more crypto/rand calls), and than filling all 16 bytes.
	src := rawBytes(dst)[n*8:]
	_, _ = rand.Read(src)

	seq := v7Seq(time.Now().UnixNano())

	// Reserve n consecutive sequence values under the lock; encoding
	// happens outside it so concurrent callers are not blocked for O(n).
	g.mu.Lock()
	seq = reserve(&g.lastSeq, seq, n)
	g.mu.Unlock()

	for i := range dst {
		copy(dst[i][8:], src[i*8:i*8+8])
		setV7(&dst[i], seq+int64(i))
	}
}
