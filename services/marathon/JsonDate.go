package marathon

import "time"

type JSONDate struct {
	*time.Time
}

func (t JSONDate) MarshalJSON() ([]byte, error) {
	return []byte(t.Format(time.RFC3339)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *JSONDate) UnmarshalJSON(data []byte) (err error) {
	// Fractional seconds are handled implicitly by Parse.
	s := string(data)
	if s == "" {
		return
	}

	tt, err := time.Parse(time.RFC3339, s)
	*t = JSONDate{&tt}
	return
}
