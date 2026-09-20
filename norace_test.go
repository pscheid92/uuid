//go:build !race

package uuid

// raceEnabled reports whether the binary was built with -race.
const raceEnabled = false
