package uuid

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"hash"
	"sync"
	"time"
)

// Pre-initialized hash states with namespace bytes already written.
// Cloned per call via hash.Cloner to avoid re-hashing the 16-byte namespace.
var (
	md5DNS  hash.Cloner
	md5URL  hash.Cloner
	md5OID  hash.Cloner
	md5X500 hash.Cloner

	sha1DNS  hash.Cloner
	sha1URL  hash.Cloner
	sha1OID  hash.Cloner
	sha1X500 hash.Cloner
)

func init() {
	md5DNS = initHash(md5.New(), NamespaceDNS)
	md5URL = initHash(md5.New(), NamespaceURL)
	md5OID = initHash(md5.New(), NamespaceOID)
	md5X500 = initHash(md5.New(), NamespaceX500)

	sha1DNS = initHash(sha1.New(), NamespaceDNS)
	sha1URL = initHash(sha1.New(), NamespaceURL)
	sha1OID = initHash(sha1.New(), NamespaceOID)
	sha1X500 = initHash(sha1.New(), NamespaceX500)
}

func initHash(h hash.Hash, ns UUID) hash.Cloner {
	h.Write(ns[:])
	return h.(hash.Cloner)
}

// NewV4 returns a new random (Version 4) UUID.
// It reads from crypto/rand which cannot fail on Go 1.26+.
func NewV4() UUID {
	var u UUID
	rand.Read(u[:])
	u[6] = (u[6] & 0x0f) | 0x40 // version 4
	u[8] = (u[8] & 0x3f) | 0x80 // variant RFC 9562
	return u
}

// NewV3 returns a deterministic Version 3 (MD5) UUID for the given namespace and name.
func NewV3(namespace UUID, name string) UUID {
	return hashUUID(namespace, name, V3, md5.New, md5DNS, md5URL, md5OID, md5X500)
}

// NewV5 returns a deterministic Version 5 (SHA-1) UUID for the given namespace and name.
func NewV5(namespace UUID, name string) UUID {
	return hashUUID(namespace, name, V5, sha1.New, sha1DNS, sha1URL, sha1OID, sha1X500)
}

// hashUUID generates a V3 or V5 UUID using the specified hash.
func hashUUID(namespace UUID, name string, ver Version, newHash func() hash.Hash, dns, url, oid, x500 hash.Cloner) UUID {
	var h hash.Hash

	// Use pre-cloned hash state for standard namespaces
	switch namespace {
	case NamespaceDNS:
		c, _ := dns.Clone()
		h = c
	case NamespaceURL:
		c, _ := url.Clone()
		h = c
	case NamespaceOID:
		c, _ := oid.Clone()
		h = c
	case NamespaceX500:
		c, _ := x500.Clone()
		h = c
	default:
		h = newHash()
		h.Write(namespace[:])
	}

	h.Write([]byte(name))
	sum := h.Sum(nil)

	var u UUID
	copy(u[:], sum[:16])
	u[6] = (u[6] & 0x0f) | (byte(ver) << 4) // version
	u[8] = (u[8] & 0x3f) | 0x80              // variant RFC 9562
	return u
}

// NewV8 returns a Version 8 UUID constructed from user-provided data.
// The version and variant bits are set; all other 122 bits come from data.
// Uniqueness is the caller's responsibility per RFC 9562 Section 5.8.
func NewV8(data [16]byte) UUID {
	u := UUID(data)
	u[6] = (u[6] & 0x0f) | 0x80 // version 8
	u[8] = (u[8] & 0x3f) | 0x80 // variant RFC 9562
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
