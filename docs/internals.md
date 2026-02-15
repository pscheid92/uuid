# Internals

This document describes how the library works under the hood. For usage, see the [README](../README.md).

## V7 Layout

A V7 UUID packs a timestamp and randomness into 128 bits:

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         unix_ts_ms (48 bits)                  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|          unix_ts_ms           | ver=7 |   rand_a (12 bits)    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|var|                      rand_b (62 bits)                     |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                          rand_b                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

- **Bytes 0-5**: 48-bit Unix timestamp in milliseconds (big-endian). This is what makes V7 UUIDs naturally sort in chronological order.
- **Bytes 6-7**: 4-bit version (`0111`) + 12-bit `rand_a` field.
- **Bytes 8-15**: 2-bit variant (`10`) + 62-bit `rand_b` from `crypto/rand`.

## V7 Sub-Millisecond Precision (RFC 9562 Method 3)

The 12-bit `rand_a` field is **not random**: it encodes sub-millisecond precision using the method from RFC 9562 Section 6.2:

```
frac = (nanoseconds_within_ms * 4096) / 1_000_000
```

This maps the 0-999,999 nanosecond range into 0-4095, giving ~244ns resolution. Combined with the 48-bit millisecond timestamp, this produces a 60-bit logical sequence:

```
seq = ms<<12 | frac
```

## V7 Monotonic Counter Fallback

When two UUIDs are generated within the same ~244ns window (same `seq` value), or when the clock hasn't advanced since the last call, the generator detects the collision and increments:

```go
if seq <= g.lastSeq {
    seq = g.lastSeq + 1  // increment to guarantee ordering
}
g.lastSeq = seq
```

The millisecond timestamp is then re-derived from the updated `seq` (`ms = seq >> 12`), so the counter can overflow into the next millisecond transparently. This means a single `Generator` can produce up to 4096 monotonically ordered UUIDs per millisecond before the timestamp advances - and continues seamlessly beyond that.

## Pool: Amortizing crypto/rand

`crypto/rand` is the dominant cost in UUID generation (~230ns per V4 call). `Pool` reduces this by pre-generating random bytes in bulk:

- **V4 pool**: Pre-stamps 256 complete UUIDs per refill (one `crypto/rand.Read` of 4KB). Each `Pool.NewV4()` call just returns the next pre-built UUID.
- **V7 pool**: Pre-generates 256 x 8-byte random chunks for `rand_b`. Timestamp and sub-ms sequence are computed live per call (they can't be pre-computed). This is why V7 pooling gives ~2x improvement vs V4's ~14x - `time.Now` is the remaining bottleneck.

## Batch: Bulk Generation

`NewV4Batch(n)` and `Generator.NewV7Batch(n)` read all random bytes in a single `crypto/rand.Read` call and stamp version/variant bits in a tight loop. For V7 batches, `time.Now` is also called once and the monotonic sequence is incremented per UUID. This avoids per-call overhead for both randomness and time, yielding ~25x (V4) and ~15x (V7) speedups over calling the single-UUID functions in a loop.

## V3/V5: hash.Cloner Optimization

V3 (MD5) and V5 (SHA-1) hash `namespace || name` to produce deterministic UUIDs. For the four standard namespaces (DNS, URL, OID, X500), the library pre-computes the hash state with the namespace bytes at init time, then uses `hash.Cloner` to clone that state per call - avoiding re-hashing the 16-byte namespace prefix every time.

## Parse: Lookup Table

Parsing uses a 256-byte hex lookup table that maps each byte value to its hex digit (or `0xFF` for invalid). Combined with a pre-computed offset array for the 32 hex character positions in `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`, this avoids branching and produces a zero-allocation parser.
