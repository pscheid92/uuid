package uuid

import (
	"testing"
	"time"
)

func BenchmarkNewV4(b *testing.B) {
	for b.Loop() {
		NewV4()
	}
}

func BenchmarkNewV5(b *testing.B) {
	for b.Loop() {
		NewV5(NamespaceDNS, "www.example.com")
	}
}

func BenchmarkNewV4Batch100(b *testing.B) {
	for b.Loop() {
		NewV4Batch(100)
	}
}

func BenchmarkNewV4Pool(b *testing.B) {
	pool := NewPool()
	for b.Loop() {
		pool.NewV4()
	}
}

func BenchmarkNewV7Pool(b *testing.B) {
	pool := NewPool()
	for b.Loop() {
		pool.NewV7()
	}
}

// BenchmarkNewV7PoolParallel measures a shared Pool under contention, where
// the length of the critical section in Pool.NewV7 bounds throughput.
func BenchmarkNewV7PoolParallel(b *testing.B) {
	pool := NewPool()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.NewV7()
		}
	})
}

func BenchmarkNewV7(b *testing.B) {
	gen := NewGenerator()
	for b.Loop() {
		gen.NewV7()
	}
}

func BenchmarkNewV7Batch100(b *testing.B) {
	gen := NewGenerator()
	for b.Loop() {
		gen.NewV7Batch(100)
	}
}

// BenchmarkNewV7BatchParallel measures a shared Generator under contention:
// every goroutine batches concurrently, so the lock scope of NewV7Batch
// directly bounds how much of the encoding can overlap.
func BenchmarkNewV7BatchParallel(b *testing.B) {
	gen := NewGenerator()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gen.NewV7Batch(1000)
		}
	})
}

// BenchmarkNewV7WithBatching measures single-UUID latency on a Generator
// that another goroutine is continuously hammering with large batches.
// The reported B/op comes from the background batcher, not from NewV7,
// which stays zero-alloc; only ns/op is meaningful here.
func BenchmarkNewV7WithBatching(b *testing.B) {
	gen := NewGenerator()
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
				gen.NewV7Batch(10000)
			}
		}
	}()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gen.NewV7()
		}
	})
	b.StopTimer()
	close(stop)
	<-done
}

func BenchmarkNewV7At(b *testing.B) {
	at := time.Now()
	for b.Loop() {
		NewV7At(at)
	}
}

func BenchmarkNewV8(b *testing.B) {
	var data [16]byte
	for b.Loop() {
		NewV8(data)
	}
}

func BenchmarkString(b *testing.B) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	for b.Loop() {
		_ = u.String()
	}
}

func BenchmarkParse(b *testing.B) {
	s := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	for b.Loop() {
		_, _ = Parse(s)
	}
}

func BenchmarkParseLenient(b *testing.B) {
	s := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	for b.Loop() {
		_, _ = ParseLenient(s)
	}
}

func BenchmarkParseLenientCompact(b *testing.B) {
	s := "6ba7b8109dad11d180b400c04fd430c8"
	for b.Loop() {
		_, _ = ParseLenient(s)
	}
}

func BenchmarkAppendText(b *testing.B) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	buf := make([]byte, 0, 36)
	for b.Loop() {
		buf = buf[:0]
		_, _ = u.AppendText(buf)
	}
}

// byteSink keeps marshaled results alive. Discarding them lets the compiler
// inline the method and keep its buffer on the stack, which reports a
// zero-alloc cost no real caller (json.Marshal, an interface call) sees.
var byteSink []byte

func BenchmarkMarshalText(b *testing.B) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	for b.Loop() {
		byteSink, _ = u.MarshalText()
	}
}

func BenchmarkUnmarshalText(b *testing.B) {
	data := []byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	var u UUID
	for b.Loop() {
		_ = u.UnmarshalText(data)
	}
}

func BenchmarkMarshalBinary(b *testing.B) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	for b.Loop() {
		byteSink, _ = u.MarshalBinary()
	}
}

func BenchmarkFromBytes(b *testing.B) {
	data := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8").Bytes()
	for b.Loop() {
		_, _ = FromBytes(data)
	}
}

func BenchmarkCompare(b *testing.B) {
	a := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	c := MustParse("6ba7b811-9dad-11d1-80b4-00c04fd430c8")
	for b.Loop() {
		Compare(a, c)
	}
}
