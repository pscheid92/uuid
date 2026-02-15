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
	mu       sync.Mutex
	lastTime int64 // last Unix millisecond timestamp used
}

// NewGenerator returns a new V7 UUID generator with its own monotonicity state.
func NewGenerator() *Generator {
	return &Generator{}
}

// NewV7 returns a new Version 7 UUID.
//
// The UUID contains a 48-bit Unix millisecond timestamp in bits 0–47,
// 12 bits of random data in bits 48–59 (with version 7 in bits 48–51),
// 2 variant bits, and 62 bits of random data in bits 66–127.
//
// When multiple UUIDs are generated within the same millisecond,
// the timestamp is incremented to guarantee monotonicity within this Generator.
func (g *Generator) NewV7() UUID {
	var u UUID
	rand.Read(u[:])

	now := time.Now()
	ms := now.UnixMilli()

	g.mu.Lock()
	if ms <= g.lastTime {
		ms = g.lastTime + 1
	}
	g.lastTime = ms
	g.mu.Unlock()

	// Encode 48-bit timestamp (big-endian) in bytes 0-5
	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)

	u[6] = (u[6] & 0x0f) | 0x70 // version 7
	u[8] = (u[8] & 0x3f) | 0x80 // variant RFC 9562
	return u
}
