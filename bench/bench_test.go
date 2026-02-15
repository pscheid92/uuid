package bench_test

import (
	"testing"

	gofrs "github.com/gofrs/uuid/v5"
	google "github.com/google/uuid"
	pscheid "github.com/pscheid92/uuid"
)

// ---------------------------------------------------------------------------
// V4 generation
// ---------------------------------------------------------------------------

func BenchmarkNewV4(b *testing.B) {
	b.Run("pscheid92", func(b *testing.B) {
		for b.Loop() {
			pscheid.NewV4()
		}
	})
	b.Run("google", func(b *testing.B) {
		for b.Loop() {
			google.New()
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		for b.Loop() {
			gofrs.NewV4()
		}
	})
}

// ---------------------------------------------------------------------------
// V7 generation
// ---------------------------------------------------------------------------

func BenchmarkNewV7(b *testing.B) {
	b.Run("pscheid92", func(b *testing.B) {
		gen := pscheid.NewGenerator()
		for b.Loop() {
			gen.NewV7()
		}
	})
	b.Run("google", func(b *testing.B) {
		for b.Loop() {
			google.NewV7()
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		for b.Loop() {
			gofrs.NewV7()
		}
	})
}

// ---------------------------------------------------------------------------
// V4 batch generation (100 UUIDs)
// ---------------------------------------------------------------------------

func BenchmarkNewV4Batch100(b *testing.B) {
	b.Run("pscheid92", func(b *testing.B) {
		for b.Loop() {
			pscheid.NewV4Batch(100)
		}
	})
	b.Run("google", func(b *testing.B) {
		for b.Loop() {
			for range 100 {
				google.New()
			}
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		for b.Loop() {
			for range 100 {
				gofrs.NewV4()
			}
		}
	})
}

// ---------------------------------------------------------------------------
// V4 pool generation (per-call amortized)
// ---------------------------------------------------------------------------

func BenchmarkNewV4Pool(b *testing.B) {
	b.Run("pscheid92", func(b *testing.B) {
		pool := pscheid.NewPool()
		for b.Loop() {
			pool.NewV4()
		}
	})
	b.Run("google", func(b *testing.B) {
		for b.Loop() {
			google.New()
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		for b.Loop() {
			gofrs.NewV4()
		}
	})
}

// ---------------------------------------------------------------------------
// V7 pool generation (per-call amortized)
// ---------------------------------------------------------------------------

func BenchmarkNewV7Pool(b *testing.B) {
	b.Run("pscheid92", func(b *testing.B) {
		pool := pscheid.NewPool()
		for b.Loop() {
			pool.NewV7()
		}
	})
	b.Run("google", func(b *testing.B) {
		for b.Loop() {
			google.NewV7()
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		for b.Loop() {
			gofrs.NewV7()
		}
	})
}

// ---------------------------------------------------------------------------
// V7 batch generation (100 UUIDs)
// ---------------------------------------------------------------------------

func BenchmarkNewV7Batch100(b *testing.B) {
	b.Run("pscheid92", func(b *testing.B) {
		gen := pscheid.NewGenerator()
		for b.Loop() {
			gen.NewV7Batch(100)
		}
	})
	b.Run("google", func(b *testing.B) {
		for b.Loop() {
			for range 100 {
				google.NewV7()
			}
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		for b.Loop() {
			for range 100 {
				gofrs.NewV7()
			}
		}
	})
}

// ---------------------------------------------------------------------------
// V5 generation (SHA-1 name-based)
// ---------------------------------------------------------------------------

func BenchmarkNewV5(b *testing.B) {
	b.Run("pscheid92", func(b *testing.B) {
		for b.Loop() {
			pscheid.NewV5(pscheid.NamespaceDNS, "www.example.com")
		}
	})
	b.Run("google", func(b *testing.B) {
		for b.Loop() {
			google.NewSHA1(google.NameSpaceDNS, []byte("www.example.com"))
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		for b.Loop() {
			gofrs.NewV5(gofrs.NamespaceDNS, "www.example.com")
		}
	})
}

// ---------------------------------------------------------------------------
// Parse (standard 36-char hyphenated form)
// ---------------------------------------------------------------------------

func BenchmarkParse(b *testing.B) {
	s := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"

	b.Run("pscheid92", func(b *testing.B) {
		for b.Loop() {
			pscheid.Parse(s)
		}
	})
	b.Run("google", func(b *testing.B) {
		for b.Loop() {
			google.Parse(s)
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		for b.Loop() {
			gofrs.FromString(s)
		}
	})
}

// ---------------------------------------------------------------------------
// String (UUID → string)
// ---------------------------------------------------------------------------

func BenchmarkString(b *testing.B) {
	b.Run("pscheid92", func(b *testing.B) {
		u := pscheid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
		for b.Loop() {
			_ = u.String()
		}
	})
	b.Run("google", func(b *testing.B) {
		u := google.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
		for b.Loop() {
			_ = u.String()
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		u := gofrs.Must(gofrs.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
		for b.Loop() {
			_ = u.String()
		}
	})
}

// ---------------------------------------------------------------------------
// MarshalText
// ---------------------------------------------------------------------------

func BenchmarkMarshalText(b *testing.B) {
	b.Run("pscheid92", func(b *testing.B) {
		u := pscheid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
		for b.Loop() {
			u.MarshalText()
		}
	})
	b.Run("google", func(b *testing.B) {
		u := google.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
		for b.Loop() {
			u.MarshalText()
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		u := gofrs.Must(gofrs.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
		for b.Loop() {
			u.MarshalText()
		}
	})
}

// ---------------------------------------------------------------------------
// UnmarshalText
// ---------------------------------------------------------------------------

func BenchmarkUnmarshalText(b *testing.B) {
	text := []byte("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	b.Run("pscheid92", func(b *testing.B) {
		var u pscheid.UUID
		for b.Loop() {
			u.UnmarshalText(text)
		}
	})
	b.Run("google", func(b *testing.B) {
		var u google.UUID
		for b.Loop() {
			u.UnmarshalText(text)
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		var u gofrs.UUID
		for b.Loop() {
			u.UnmarshalText(text)
		}
	})
}

// ---------------------------------------------------------------------------
// MarshalBinary
// ---------------------------------------------------------------------------

func BenchmarkMarshalBinary(b *testing.B) {
	b.Run("pscheid92", func(b *testing.B) {
		u := pscheid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
		for b.Loop() {
			u.MarshalBinary()
		}
	})
	b.Run("google", func(b *testing.B) {
		u := google.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
		for b.Loop() {
			u.MarshalBinary()
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		u := gofrs.Must(gofrs.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
		for b.Loop() {
			u.MarshalBinary()
		}
	})
}

// ---------------------------------------------------------------------------
// UnmarshalBinary
// ---------------------------------------------------------------------------

func BenchmarkUnmarshalBinary(b *testing.B) {
	data := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}

	b.Run("pscheid92", func(b *testing.B) {
		var u pscheid.UUID
		for b.Loop() {
			u.UnmarshalBinary(data)
		}
	})
	b.Run("google", func(b *testing.B) {
		var u google.UUID
		for b.Loop() {
			u.UnmarshalBinary(data)
		}
	})
	b.Run("gofrs", func(b *testing.B) {
		var u gofrs.UUID
		for b.Loop() {
			u.UnmarshalBinary(data)
		}
	})
}
