package uuid

import (
	"crypto/rand"
	"sync"
	"time"
)

// defaultGen is the package-level V7 generator, analogous to http.DefaultClient.
var defaultGen = NewGenerator()

// NewV7 returns a new Version 7 (Unix timestamp + random) UUID using the
// package-level default generator. For isolated monotonicity guarantees,
// create a dedicated [Generator] with [NewGenerator].
func NewV7() UUID {
	return defaultGen.NewV7()
}

// Generator produces Version 7 UUIDs with per-instance monotonicity.
// Multiple goroutines may safely call NewV7 concurrently on the same Generator.
type Generator struct {
	mu      sync.Mutex
	lastSeq int64 // ms<<12 | seq for monotonicity
}

// NewGenerator returns a new V7 UUID generator with its own monotonicity state.
func NewGenerator() *Generator {
	return &Generator{}
}

const nanoPerMilli = 1_000_000

// NewV7 returns a new Version 7 UUID.
//
// The UUID encodes a 48-bit Unix millisecond timestamp in bits 0–47 and
// 12 bits of sub-millisecond precision in the rand_a field (bits 48–59),
// computed per RFC 9562 Section 6.2 Method 3. The rand_b field (bits 66–127)
// is filled with random data from crypto/rand.
//
// When multiple UUIDs are generated faster than the clock resolution,
// the combined timestamp+seq counter is incremented to guarantee
// monotonicity within this Generator.
func (g *Generator) NewV7() UUID {
	var u UUID
	rand.Read(u[:])

	now := time.Now()
	nano := now.UnixNano()
	ms := nano / nanoPerMilli
	// RFC 9562 Section 6.2 Method 3: sub-millisecond precision scaled to 12 bits.
	frac := (nano % nanoPerMilli) * 4096 / nanoPerMilli
	seq := ms<<12 | frac

	g.mu.Lock()
	if seq <= g.lastSeq {
		seq = g.lastSeq + 1
	}
	g.lastSeq = seq
	g.mu.Unlock()

	ms = seq >> 12
	seq12 := seq & 0xFFF

	// Encode 48-bit timestamp (big-endian) in bytes 0-5
	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)

	// Encode version 7 and 12-bit sub-millisecond sequence in bytes 6-7
	u[6] = 0x70 | byte(seq12>>8)&0x0f
	u[7] = byte(seq12)

	u[8] = (u[8] & 0x3f) | 0x80 // variant RFC 9562
	return u
}
