package kafka

import (
	"fmt"
	"regexp"
	"strconv"
)

func convertDurationToMilliseconds(d string) (int, error) {
	// Use regex to parse the duration string value.
	// The parsing should return an index length of 3. Example: [10d 10 d]
	regex := regexp.MustCompile(`^(\d+) ?(ms|[smhdw]|milliseconds?|seconds?|minutes?|hours?|days?|weeks?)$`)
	result := regex.FindStringSubmatch(d)

	if len(result) != 3 {
		return 0, fmt.Errorf("unknown duration string")
	}

	multiplier := 1
	magnitude, _ := strconv.Atoi(result[1])
	unit := result[2]

	switch unit {
	case "ms", "millisecond", "milliseconds":
		multiplier = 1
	case "s", "second", "seconds":
		multiplier = 1000
	case "m", "minute", "minutes":
		multiplier = 1000 * 60
	case "h", "hour", "hours":
		multiplier = 1000 * 60 * 60
	case "d", "day", "days":
		multiplier = 1000 * 60 * 60 * 24
	case "w", "week", "weeks":
		multiplier = 1000 * 60 * 60 * 24 * 7
	default:
		return 0, nil
	}

	return multiplier * magnitude, nil
}
