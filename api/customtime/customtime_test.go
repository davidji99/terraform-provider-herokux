package customtime

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomTimeJSONRoundTrip(t *testing.T) {
	for _, input := range []string{`"2024-07-31 01:15:11 +0000"`, `"2024-07-31 01:15:11 +0900"`} {
		t.Run(input, func(t *testing.T) {
			var parsed CustomTime
			if err := json.Unmarshal([]byte(input), &parsed); err != nil {
				t.Fatal(err)
			}

			encoded, err := json.Marshal(&parsed)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, input, string(encoded))
		})
	}
}
