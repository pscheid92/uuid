//go:build race

package uuid

// raceEnabled reports whether the binary was built with -race.
// Zero-alloc tests skip under the race detector because crypto/rand.Read
// allocates once per call there on Linux (observed with Go 1.27 on
// linux/amd64 and linux/arm64), which is a property of the instrumented
// runtime rather than of this package.
const raceEnabled = true
