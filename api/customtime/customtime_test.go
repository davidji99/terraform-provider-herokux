package customtime

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestArgs struct {
	Time CustomTime
}

func TestHelloName_UTC(t *testing.T) {
	data := `
    {"Time": "2024-07-31 01:15:11 +0000"}
`

	a := TestArgs{}
	unmarshalErr := json.Unmarshal([]byte(data), &a)
	assert.Nil(t, unmarshalErr)
	assert.Equal(t, "2024-07-31 01:15:11 +0000 +0000", a.Time.String())

	marshalledData, marshallErr := json.Marshal(a.Time.String())
	assert.Nil(t, marshallErr)
	assert.Equal(t, "2024-07-31 01:15:11 +0000 +0000", string(marshalledData))
}

func TestHelloName_JST(t *testing.T) {
	data := `
    {"Time": "2024-07-31 01:15:11 +0900"}
`

	a := TestArgs{}
	unmarshalErr := json.Unmarshal([]byte(data), &a)
	assert.Nil(t, unmarshalErr)
	assert.Equal(t, "2024-07-31 01:15:11 +0900 JST", a.Time.String())

	marshalledData, marshallErr := json.Marshal(a.Time.String())
	assert.Nil(t, marshallErr)
	assert.Equal(t, "2024-07-31 01:15:11 +0900 JST", string(marshalledData))
}
