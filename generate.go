package uuid

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"hash"
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
	return hashUUID(namespace, name, Version3, md5.New, md5DNS, md5URL, md5OID, md5X500)
}

// NewV5 returns a deterministic Version 5 (SHA-1) UUID for the given namespace and name.
func NewV5(namespace UUID, name string) UUID {
	return hashUUID(namespace, name, Version5, sha1.New, sha1DNS, sha1URL, sha1OID, sha1X500)
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
