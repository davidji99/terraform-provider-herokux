package kafka

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvertDurationToMilliseconds_MS_Valid(t *testing.T) {
	expected := 2454665

	for _, d := range []string{"2454665ms", "2454665millisecond", "2454665milliseconds"} {
		v, err := ConvertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_S_Valid(t *testing.T) {
	expected := 9845000

	for _, d := range []string{"9845s", "9845second", "9845seconds"} {
		v, err := ConvertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_M_Valid(t *testing.T) {
	expected := 590700000

	for _, d := range []string{"9845m", "9845minute", "9845minutes"} {
		v, err := ConvertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_H_Valid(t *testing.T) {
	expected := 277200000

	for _, d := range []string{"77h", "77hour", "77hours"} {
		v, err := ConvertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_D_Valid(t *testing.T) {
	expected := 1209600000

	for _, d := range []string{"14d", "14day", "14days"} {
		v, err := ConvertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_W_Valid(t *testing.T) {
	expected := 1814400000

	for _, d := range []string{"3w", "3week", "3weeks"} {
		v, err := ConvertDurationToMilliseconds(d)
		assert.Nil(t, err)
		assert.Equal(t, expected, v)
	}
}

func TestConvertDurationToMilliseconds_Invalid(t *testing.T) {
	testDuration := "1month"
	v, err := ConvertDurationToMilliseconds(testDuration)
	assert.NotNil(t, err)
	assert.Equal(t, 0, v)

}

func TestConvertMStoDuration_Week(t *testing.T) {
	expected := "2w"
	testMS := 1209600000

	result, err := ConvertMillisecondsToDuration(testMS)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestConvertMStoDuration_Day(t *testing.T) {
	expected := "8d"
	testMS := 691200000

	result, err := ConvertMillisecondsToDuration(testMS)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestConvertMStoDuration_Hour(t *testing.T) {
	expected := "47h"
	testMS := 169200000

	result, err := ConvertMillisecondsToDuration(testMS)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestConvertMStoDuration_Minute(t *testing.T) {
	expected := "123m"
	testMS := 7380000

	result, err := ConvertMillisecondsToDuration(testMS)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestConvertMStoDuration_Second(t *testing.T) {
	expected := "123s"
	testMS := 123000

	result, err := ConvertMillisecondsToDuration(testMS)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestConvertMStoDuration_Millisecond(t *testing.T) {
	expected := "12300ms"
	testMS := 12300

	result, err := ConvertMillisecondsToDuration(testMS)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestConvertMStoDuration_HigherDurationWins(t *testing.T) {
	expected := "1w"
	testMS := 604800000

	result, err := ConvertMillisecondsToDuration(testMS)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}
