package triax

import (
	"bytes"
	"encoding/json"
)

// quotedInt is an integer wrapped in quotes. For whatever reason,
// the controller sometimes wraps integer values in quotes.
type quotedInt int

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (i *quotedInt) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		return nil
	}

	if l := len(b); l > 2 && b[0] == '"' && b[l-1] == '"' {
		b = b[1 : l-1]
	}

	var val int
	if err := json.Unmarshal(b, &val); err != nil {
		return err //nolint:wrapcheck
	}

	*i = quotedInt(val)
	return nil
}
