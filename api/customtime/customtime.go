package customtime

import (
	"encoding/json"
	"time"
)

type CustomTime struct {
	time.Time
}

const ctLayout = "2006/01/02|15:04:05"

func (ct *CustomTime) UnmarshalJSON(data []byte) (err error) {
	// Remove quotes from JSON string
	str := string(data[1 : len(data)-1])

	// Parse the custom format: "2024-07-31 01:16:17 +0000"
	t, err := time.Parse("2006-01-02 15:04:05 -0700", str)
	if err != nil {
		return err
	}

	ct.Time = t
	return nil

}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(ct.Time.Format("2006-01-02 15:04:05 -0700"))
}

var nilTime = (time.Time{}).UnixNano()

func (ct *CustomTime) IsSet() bool {
	return !ct.IsZero()
}

type Args struct {
	Time CustomTime
}
