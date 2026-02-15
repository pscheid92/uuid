package uuid_test

import (
	"fmt"
	"slices"

	"github.com/pscheid92/uuid"
)

func ExampleNewV4() {
	id := uuid.NewV4()
	fmt.Println(id.Version())
	// Output: V4
}

func ExampleNewV3() {
	id := uuid.NewV3(uuid.NamespaceDNS, "www.example.com")
	fmt.Println(id)
	// Output: 5df41881-3aed-3515-88a7-2f4a814cf09e
}

func ExampleNewV5() {
	id := uuid.NewV5(uuid.NamespaceDNS, "www.example.com")
	fmt.Println(id)
	// Output: 2ed6657d-e927-568b-95e1-2665a8aea6a2
}

func ExampleNewV7() {
	id := uuid.NewV7()
	fmt.Println(id.Version())
	// Output: V7
}

func ExampleNewV8() {
	var data [16]byte
	copy(data[:], "custom-data-here")
	id := uuid.NewV8(data)
	fmt.Println(id.Version())
	// Output: V8
}

func ExampleParse() {
	id, err := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	// Output: 6ba7b810-9dad-11d1-80b4-00c04fd430c8
}

func ExampleParseLenient() {
	// Accepts URN, braced, and compact forms in addition to standard
	id, err := uuid.ParseLenient("urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	// Output: 6ba7b810-9dad-11d1-80b4-00c04fd430c8
}

func ExampleUUID_URN() {
	id := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	fmt.Println(id.URN())
	// Output: urn:uuid:6ba7b810-9dad-11d1-80b4-00c04fd430c8
}

func ExampleCompare() {
	ids := []uuid.UUID{
		uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		uuid.MustParse("00000000-0000-0000-0000-000000000002"),
	}
	slices.SortFunc(ids, uuid.Compare)
	for _, id := range ids {
		fmt.Println(id)
	}
	// Output:
	// 00000000-0000-0000-0000-000000000001
	// 00000000-0000-0000-0000-000000000002
	// 00000000-0000-0000-0000-000000000003
}

func ExampleGenerator() {
	gen := uuid.NewGenerator()
	id := gen.NewV7()
	fmt.Println(id.Version())
	// Output: V7
}

func ExampleFromBytes() {
	b := []byte{0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1, 0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8}
	id, err := uuid.FromBytes(b)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	// Output: 6ba7b810-9dad-11d1-80b4-00c04fd430c8
}
