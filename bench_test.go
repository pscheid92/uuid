package uuid

import "testing"

func BenchmarkNewV4(b *testing.B) {
	for b.Loop() {
		NewV4()
	}
}

func BenchmarkNewV3(b *testing.B) {
	for b.Loop() {
		NewV3(NamespaceDNS, "www.example.com")
	}
}

func BenchmarkNewV5(b *testing.B) {
	for b.Loop() {
		NewV5(NamespaceDNS, "www.example.com")
	}
}

func BenchmarkNewV7(b *testing.B) {
	gen := NewGenerator()
	for b.Loop() {
		gen.NewV7()
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
		Parse(s)
	}
}

func BenchmarkParseLenient(b *testing.B) {
	s := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	for b.Loop() {
		ParseLenient(s)
	}
}

func BenchmarkParseLenientCompact(b *testing.B) {
	s := "6ba7b8109dad11d180b400c04fd430c8"
	for b.Loop() {
		ParseLenient(s)
	}
}

func BenchmarkAppendText(b *testing.B) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	buf := make([]byte, 0, 36)
	for b.Loop() {
		buf = buf[:0]
		u.AppendText(buf)
	}
}

func BenchmarkMarshalText(b *testing.B) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	for b.Loop() {
		u.MarshalText()
	}
}

func BenchmarkMarshalBinary(b *testing.B) {
	u := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	for b.Loop() {
		u.MarshalBinary()
	}
}

func BenchmarkFromBytes(b *testing.B) {
	data := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8").Bytes()
	for b.Loop() {
		FromBytes(data)
	}
}

func BenchmarkCompare(b *testing.B) {
	a := MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	c := MustParse("6ba7b811-9dad-11d1-80b4-00c04fd430c8")
	for b.Loop() {
		Compare(a, c)
	}
}
