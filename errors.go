package uuid

import "fmt"

// ParseError is returned when a UUID string cannot be parsed.
//
// Use [errors.AsType] to check for this error:
//
//	if perr, ok := errors.AsType[*ParseError](err); ok {
//	    fmt.Println(perr.Input)
//	}
type ParseError struct {
	Input string // the string that failed to parse
	Msg   string // description of the problem
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("uuid: parsing %q: %s", e.Input, e.Msg)
}

// LengthError is returned when the input has an unexpected byte length.
//
// Use [errors.AsType] to check for this error:
//
//	if lerr, ok := errors.AsType[*LengthError](err); ok {
//	    fmt.Println(lerr.Got, lerr.Want)
//	}
type LengthError struct {
	Got  int    // the actual length
	Want string // description of expected length
}

func (e *LengthError) Error() string {
	return fmt.Sprintf("uuid: unexpected length %d, want %s", e.Got, e.Want)
}
