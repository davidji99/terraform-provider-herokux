package kafka

import (
	"fmt"
	"github.com/elliotchance/orderedmap/v2"
	"regexp"
	"strconv"
)

const (
	// RetentionTimeDurationRegex is the regex for the string form of the duration,
	RetentionTimeDurationRegex = `^(\d+) ?(ms|[smhdw]|milliseconds?|seconds?|minutes?|hours?|days?|weeks?)$`

	// RetentionTimeDurationRegexStricter is the same as RetentionTimeDurationRegex
	// but only allows for duration abbreviations.
	RetentionTimeDurationRegexStricter = `^(\d+) ?(ms|[smhdw])$`

	// RententionTimeDuragionRegexStricterWithDisable is the same as RetentionTimeDurationRegexStricter but allows
	// 'disable' as a value in order to disable retention time.
	RetentionTimeDuragionRegexStricterWithDisable = `^(\d+) ?(ms|[smhdw]|)|disable$`

	RetentionTimeDisableVal = "disable"

	// Multiplier constants
	WeekMultiplier        = DayMultiplier * 7
	DayMultiplier         = HourMultiplier * 24
	HourMultiplier        = MinuteMultiplier * 60
	MinuteMultiplier      = SecondMultiplier * 60
	SecondMultiplier      = MillisecondMultiplier * 1000
	MillisecondMultiplier = 1
)

// ConvertDurationToMilliseconds converts a duration unit to milliseconds integer value.
//
// Example: "10d". Supported suffixes:
//   - `ms`, `millisecond`, `milliseconds`
//   - `s`, `second`, `seconds`
//   - `m`, `minute`, `minutes`
//   - `h`, `hour`, `hours`
//   - `d`, `day`, `days`
//   - `w`, `week`, `weeks`
func ConvertDurationToMilliseconds(d string) (int, error) {
	// Use regex to parse the duration string value.
	// The parsing should return an index length of 3. Example: [10d 10 d]
	regex := regexp.MustCompile(RetentionTimeDurationRegex)
	result := regex.FindStringSubmatch(d)

	if len(result) != 3 {
		return 0, fmt.Errorf("unknown duration string")
	}

	var multiplier int
	magnitude, _ := strconv.Atoi(result[1])
	unit := result[2]

	switch unit {
	case "ms", "millisecond", "milliseconds":
		multiplier = MillisecondMultiplier
	case "s", "second", "seconds":
		multiplier = SecondMultiplier
	case "m", "minute", "minutes":
		multiplier = MinuteMultiplier
	case "h", "hour", "hours":
		multiplier = HourMultiplier
	case "d", "day", "days":
		multiplier = DayMultiplier
	case "w", "week", "weeks":
		multiplier = WeekMultiplier
	default:
		return 0, nil
	}

	return multiplier * magnitude, nil
}

// ConvertMillisecondsToDuration takes a millisecond integer parameter and converts it to a duration string
// in the format of "<number><ms|s|m|h|d|w<".
//
// If the millisecond parameter overlaps with two durations such as 1w or 7d,
// this function returns the longer duration unit.
func ConvertMillisecondsToDuration(ms int) (string, error) {
	// Loop through multiplier in order of largest to smallest. Return the first pair that divides cleanly.
	for el := multipliersMap().Front(); el != nil; el = el.Next() {
		duration := el.Key
		multi := el.Value
		if ms%multi == 0 {
			return fmt.Sprintf("%d%s", ms/multi, duration), nil
		}
	}

	return "", fmt.Errorf("unable to parse milliseconds to duration")
}

// multipliersMap is an ordered map of longest to shortest duration unit and their respective multipliers.
func multipliersMap() *orderedmap.OrderedMap[string, int] {
	m := orderedmap.NewOrderedMap[string, int]()

	m.Set("w", WeekMultiplier)
	m.Set("d", DayMultiplier)
	m.Set("h", HourMultiplier)
	m.Set("m", MinuteMultiplier)
	m.Set("s", SecondMultiplier)
	m.Set("ms", MillisecondMultiplier)

	return m
}
