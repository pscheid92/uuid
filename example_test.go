package uuid_test

import (
	"database/sql/driver"
	"fmt"
	"slices"
	"time"
	stdlib "uuid"

	"github.com/pscheid92/uuid"
)

func ExampleNewV4() {
	id := uuid.NewV4()
	fmt.Println(id.Version())
	// Output: V4
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

func ExampleNewV7At() {
	// Backfill a record that was created before V7 keys were introduced.
	created := time.Date(2020, time.March, 14, 15, 9, 26, 0, time.UTC)
	id := uuid.NewV7At(created)

	ts, ok := id.Time()
	fmt.Println(id.Version(), ok, ts.UTC())
	// Output: V7 true 2020-03-14 15:09:26 +0000 UTC
}

func ExampleUUID_Time() {
	v7 := uuid.NewV7At(time.UnixMilli(1_700_000_000_000))
	v4 := uuid.NewV4()

	_, ok := v7.Time()
	fmt.Println("V7 has a timestamp:", ok)
	_, ok = v4.Time()
	fmt.Println("V4 has a timestamp:", ok)
	// Output:
	// V7 has a timestamp: true
	// V4 has a timestamp: false
}

func ExampleNewV4Batch() {
	ids := uuid.NewV4Batch(3)
	fmt.Println(len(ids))
	fmt.Println(ids[0].Version())
	// Output:
	// 3
	// V4
}

func ExamplePool_NewV4() {
	pool := uuid.NewPool()
	id := pool.NewV4()
	fmt.Println(id.Version())
	// Output: V4
}

func ExamplePool_NewV7() {
	pool := uuid.NewPool()
	id := pool.NewV7()
	fmt.Println(id.Version())
	// Output: V7
}

func ExampleGenerator_NewV7Batch() {
	gen := uuid.NewGenerator()
	ids := gen.NewV7Batch(3)
	fmt.Println(len(ids))
	fmt.Println(ids[0].Version())
	// Output:
	// 3
	// V7
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

// BinaryUUID stores the UUID as 16 raw bytes, for BINARY(16) columns
// (MySQL, MariaDB). Scan already accepts 16 raw bytes, so only Value
// needs to be overridden; every other method is promoted from uuid.UUID.
type BinaryUUID struct{ uuid.UUID }

func (b BinaryUUID) Value() (driver.Value, error) {
	return b.Bytes(), nil
}

func ExampleUUID_Value() {
	id := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	v, _ := id.Value() // string, for native uuid column types
	fmt.Printf("%T %v\n", v, v)

	bv, _ := BinaryUUID{id}.Value() // []byte, for BINARY(16)
	fmt.Printf("%T %d bytes\n", bv, len(bv.([]byte)))
	// Output:
	// string 6ba7b810-9dad-11d1-80b4-00c04fd430c8
	// []uint8 16 bytes
}

// Both this package and the standard library uuid package (Go 1.27+) define
// UUID as [16]byte, so values convert in either direction at zero cost.
func ExampleUUID_stdlibInterop() {
	std := stdlib.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	// Standard library -> this package: gain Version, Time, Scan/Value, ...
	id := uuid.UUID(std)
	fmt.Println(id.Version())

	// This package -> standard library.
	back := stdlib.UUID(uuid.NewV5(uuid.NamespaceDNS, "example.com"))
	fmt.Println(back == stdlib.UUID(uuid.NewV5(uuid.NamespaceDNS, "example.com")))
	// Output:
	// V1
	// true
}
