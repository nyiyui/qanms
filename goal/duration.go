package goal

import (
	"encoding/json"
	"time"
)

// Duration is a encoding-friendly time.Duration.
type Duration time.Duration

// UnmarshalJSON implements json.Unmarshaler.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var raw string
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	d2, err := time.ParseDuration(raw)
	*d = Duration(d2)
	return err
}

// MarshalJSON implements json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}
