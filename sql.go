package uuid

import (
	"database/sql/driver"
	"fmt"
)

// Scan implements [database/sql.Scanner]. It supports scanning from:
//   - string: parsed with [ParseLenient]
//   - []byte: 16 raw bytes or text form parsed with [ParseLenient]
//
// For SQL NULL handling, use *UUID (nil pointer = NULL).
func (u *UUID) Scan(src any) error {
	switch v := src.(type) {
	case string:
		parsed, err := ParseLenient(v)
		if err != nil {
			return err
		}
		*u = parsed
		return nil

	case []byte:
		if len(v) == 16 {
			copy(u[:], v)
			return nil
		}
		parsed, err := ParseLenient(string(v))
		if err != nil {
			return err
		}
		*u = parsed
		return nil

	default:
		return fmt.Errorf("uuid: cannot scan %T into UUID", src)
	}
}

// Value implements [database/sql/driver.Valuer].
// It returns the UUID as a 36-character string.
func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}
