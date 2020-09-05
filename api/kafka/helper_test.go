package kafka

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertDurationToMilliseconds_MS_Valid(t *testing.T) {
	expected := 2454665

	for _, d := range []string{"2454665ms", "2454665millisecond", "2454665milliseconds"} {
		v, err := convertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_S_Valid(t *testing.T) {
	expected := 9845000

	for _, d := range []string{"9845s", "9845second", "9845seconds"} {
		v, err := convertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_M_Valid(t *testing.T) {
	expected := 590700000

	for _, d := range []string{"9845m", "9845minute", "9845minutes"} {
		v, err := convertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_H_Valid(t *testing.T) {
	expected := 277200000

	for _, d := range []string{"77h", "77hour", "77hours"} {
		v, err := convertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_D_Valid(t *testing.T) {
	expected := 1209600000

	for _, d := range []string{"14d", "14day", "14days"} {
		v, err := convertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_W_Valid(t *testing.T) {
	expected := 1814400000

	for _, d := range []string{"3w", "3week", "3weeks"} {
		v, err := convertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_Invalid(t *testing.T) {
	testDuration := "1month"
	v, err := convertDurationToMilliseconds(testDuration)
	assert.NotNil(t, err)
	assert.Equal(t, 0, v)

}
