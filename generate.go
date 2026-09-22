package uuid

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"io"
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
// in NewV5 without allocating. Names longer than v5StackBuf-16 bytes fall
// back to a streaming hash.
const v5StackBuf = 256

// NewV5 returns a deterministic Version 5 (SHA-1) UUID for the given namespace and name.
// It allocates nothing for names up to 240 bytes.
func NewV5(namespace UUID, name string) UUID {
	var sum [sha1.Size]byte
	if len(name) <= v5StackBuf-len(namespace) {
		// Hash namespace||name from a stack buffer in one call.
		var buf [v5StackBuf]byte
		copy(buf[:], namespace[:])
		n := copy(buf[len(namespace):], name)
		sum = sha1.Sum(buf[:len(namespace)+n])
	} else {
		h := sha1.New()
		h.Write(namespace[:])
		_, _ = io.WriteString(h, name)
		h.Sum(sum[:0])
	}

	var u UUID
	copy(u[:], sum[:16])
	stamp(&u, V5)
	return u
}

// NewV4Batch returns n random (Version 4) UUIDs.
// It amortizes the cost of crypto/rand by reading all random bytes in a
// single call, making it significantly faster than calling [NewV4] in a loop.
// It returns nil if n <= 0.
func NewV4Batch(n int) []UUID {
	if n <= 0 {
		return nil
	}
	uuids := make([]UUID, n)
	_, _ = rand.Read(rawBytes(uuids))
	for i := range uuids {
		stamp(&uuids[i], V4)
	}
	return uuids
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
	v7left int   // 8-byte chunks remaining in v7rand; zero triggers a refill
	v7seq  int64 // ms<<12 | seq for V7 monotonicity
}

const poolSize = 256

// NewPool returns a new [Pool] that amortizes crypto/rand overhead.
func NewPool() *Pool {
	return &Pool{}
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
// It is functionally equivalent to [Generator.NewV7] but amortizes
// the crypto/rand overhead by buffering random bytes for the rand_b field.
// Timestamps are computed live to remain accurate, though under sustained
// bursts they may run slightly ahead of the wall clock (see [Generator.NewV7]).
//
// Each Pool keeps its own monotonic state, independent of the package-level
// [NewV7] generator and of every other Pool or [Generator]. UUIDs drawn from
// different sources are not ordered relative to each other.
func (p *Pool) NewV7() UUID {
	// Read the clock before taking the lock, as Generator.NewV7 does, so
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
	seq = reserve(&p.v7seq, seq, 1)
	p.mu.Unlock()

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
// It amortizes the cost of crypto/rand and [time.Now] by performing a single
// call of each, making it significantly faster than calling [Generator.NewV7]
// in a loop. It returns nil if n <= 0.
//
// All n UUIDs derive from one clock reading: consecutive counter values can
// carry into the millisecond field, so for large n the encoded timestamps
// may run slightly ahead of the wall clock.
func (g *Generator) NewV7Batch(n int) []UUID {
	if n <= 0 {
		return nil
	}
	uuids := make([]UUID, n)

	// One bulk random read for all rand_b fields. Reading only the 8
	// random bytes per UUID into a side buffer is faster than filling the
	// whole result and overwriting bytes 0–7.
	randBuf := make([]byte, n*8)
	_, _ = rand.Read(randBuf)

	seq := v7Seq(time.Now().UnixNano())

	// Reserve n consecutive sequence values under the lock; encoding
	// happens outside it so concurrent callers are not blocked for O(n).
	g.mu.Lock()
	seq = reserve(&g.lastSeq, seq, n)
	g.mu.Unlock()

	for i := range uuids {
		copy(uuids[i][8:], randBuf[i*8:i*8+8])
		setV7(&uuids[i], seq+int64(i))
	}
	return uuids
}
