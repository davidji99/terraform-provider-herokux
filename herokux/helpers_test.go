package herokux

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestParseFrequency_EveryMinute(t *testing.T) {
	every, at, err := parseFrequency("every_ten_minutes")

	assert.Nil(t, err)
	assert.Equal(t, 10, every)
	assert.Equal(t, 0, at)
}

func TestParseFrequency_EveryHour(t *testing.T) {
	cadence := []int{0, 10, 20, 30, 40, 50}

	for _, c := range cadence {
		every, at, err := parseFrequency(fmt.Sprintf("every_hour_at_%d", c))
		assert.Nil(t, err)
		assert.Equal(t, 60, every)
		assert.Equal(t, c, at)
	}
}

func TestParseFrequency_EveryDay(t *testing.T) {
	hours := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14",
		"15", "16", "17", "18", "19", "20", "21", "22", "23"}
	minutes := []string{"30", "00"}

	for _, h := range hours {
		for _, m := range minutes {
			every, at, err := parseFrequency(fmt.Sprintf("every_day_at_%s:%s", h, m))
			assert.Nil(t, err)
			assert.Equal(t, 1440, every)

			hourInt, _ := strconv.Atoi(h)
			minuteInt, _ := strconv.Atoi(m)
			assert.Equal(t, (hourInt*60)+minuteInt, at)
		}
	}
}
