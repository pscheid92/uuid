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

When two UUIDs are generated within the same ~244ns window (same `seq` value), when the clock hasn't advanced since the last call, or when it has stepped back, the generator continues from the last value instead. `Generator.NewV7`, `Pool.NewV7`, and `NewV7Batch` all reserve their values through one helper, called under the generator's lock (`n` is 1 for a single UUID):

```go
func reserve(last *int64, seq int64, n int) int64 {
    if seq <= *last {
        seq = *last + 1 // continue after the last value to guarantee ordering
    }
    *last = seq + int64(n-1)
    return seq
}
```

The millisecond timestamp is then re-derived from the updated `seq` (`ms = seq >> 12`), so the counter can overflow into the next millisecond transparently. The counter starts from the current sub-millisecond fraction, so a burst can take at most 4096 - `frac` values before it carries into the next millisecond; ordering continues seamlessly past that. The cost is that the encoded timestamp then runs ahead of the wall clock until real time catches up, which is why `UUID.Time` may report a slightly later time under sustained bursts.

## Pool: Amortizing crypto/rand

`crypto/rand` is the dominant cost in UUID generation (~240ns per V4 call). `Pool` reduces this by pre-generating random bytes in bulk:

- **V4 pool**: Pre-stamps 256 complete UUIDs per refill (one `crypto/rand.Read` of 4KB). Each `Pool.NewV4()` call just returns the next pre-built UUID.
- **V7 pool**: Pre-generates 256 x 8-byte random chunks for `rand_b`. Timestamp and sub-ms sequence are computed live per call (they can't be pre-computed). This is why V7 pooling gives ~2x improvement vs V4's ~14x - `time.Now` is the remaining bottleneck.

## Batch: Bulk Generation

`NewV4Batch(n)` and `Generator.NewV7Batch(n)` read all random bytes in a single `crypto/rand.Read` call and stamp version/variant bits in a tight loop. For V7 batches, `time.Now` is also called once and the monotonic sequence is incremented per UUID. This avoids per-call overhead for both randomness and time, yielding ~30x (V4) and ~13x (V7) speedups at n=100 over calling the single-UUID functions in a loop, growing with n (~40x and ~20x at n=1000).

## V5: Zero-Alloc Hashing

V5 (SHA-1) hashes `namespace || name` to produce deterministic UUIDs. For names up to 240 bytes, the library concatenates namespace and name into a 256-byte stack buffer and calls `sha1.Sum` directly, which takes a concrete `[]byte` and returns a `[20]byte` - nothing escapes to the heap, so `NewV5` is zero-alloc. Longer names fall back to the streaming `hash.Hash` API, which allocates because arguments passed through an interface escape.

This replaced an earlier `hash.Cloner` approach that pre-hashed the namespace: since a 16-byte namespace never fills a 64-byte SHA-1 block, cloning saved no compression rounds, and the interface calls cost four allocations per UUID.

## Batch: Filling the Result Directly

`NewV4Batch` lets `crypto/rand` write straight into the `[]UUID` backing array via `unsafe.Slice`, so the only allocation is the result itself. A `[]UUID` is a contiguous array of `[16]byte`, so the reinterpretation is exact. `NewV7Batch` keeps a separate 8-byte-per-UUID buffer for `rand_b`: reading half as many random bytes measured faster than filling all 16 and overwriting the timestamp.

## Parse: Lookup Table

Parsing uses a 256-byte hex lookup table that maps each byte value to its hex digit (or `0xFF` for invalid). Combined with a pre-computed array of the offsets of the 16 hex digit pairs in `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`, this replaces per-character range comparisons with a single table lookup and produces a zero-allocation parser.
