package metrics

// FormationMonitorName represents a formation monitor name.
type FormationMonitorName string

// FormationMonitorNames represents all the names available.
var FormationMonitorNames = struct {
	// Latency is the alert for monitoring response times.
	Latency FormationMonitorName

	// LatencyScale is the name for the autoscaling configuration.
	LatencyScale FormationMonitorName

	// ErrorRate is the alert for monitoring failed requests.
	ErrorRate FormationMonitorName
}{
	Latency:      "LATENCY",
	LatencyScale: "LATENCY_SCALE",
	ErrorRate:    "ERROR_RATE",
}

// ToString is a helper method to return the string of a FormationMonitorName.
func (f FormationMonitorName) ToString() string {
	return string(f)
}

// FormationMonitorActionType represents a formation monitor's action type.
type FormationMonitorActionType string

// FormationMonitorActionTypes represents all possible action types.
var FormationMonitorActionTypes = struct {
	Alert FormationMonitorActionType
	Scale FormationMonitorActionType
}{
	Alert: "alert",
	Scale: "scale",
}

// ToString is a helper method to return the string of a FormationMonitorActionType.
func (f FormationMonitorActionType) ToString() string {
	return string(f)
}

// ThresholdAlertEmailFrequency represents the alert email frequency.
type ThresholdAlertEmailFrequency string

// ThresholdAlertEmailFrequencies represents all available alert frequencies.
var ThresholdAlertEmailFrequencies = struct {
	EveryFiveMinutes ThresholdAlertEmailFrequency
	EveryHour        ThresholdAlertEmailFrequency
	EveryDay         ThresholdAlertEmailFrequency
}{
	EveryFiveMinutes: "every_five_minutes",
	EveryHour:        "every_hour",
	EveryDay:         "every_day",
}

// ToString is a helper method to return the string of a ThresholdAlertEmailFrequency.
func (t ThresholdAlertEmailFrequency) ToString() string {
	return string(t)
}

// ToInt is a helper method to converts the ThresholdAlertEmailFrequency to its integer value.
func (t ThresholdAlertEmailFrequency) ToInt() int {
	switch t {
	case ThresholdAlertEmailFrequencies.EveryFiveMinutes:
		return 5
	case ThresholdAlertEmailFrequencies.EveryHour:
		return 60
	case ThresholdAlertEmailFrequencies.EveryDay:
		return 1440
	default:
		return 5
	}
}

// ThresholdAlertSensitivity represents the alert sensitivity.
type ThresholdAlertSensitivity string

// ThresholdAlertSensitives represents all available alert sensitivities.
var ThresholdAlertSensitives = struct {
	High   ThresholdAlertSensitivity
	Medium ThresholdAlertSensitivity
	Low    ThresholdAlertSensitivity
}{
	High:   "high",
	Medium: "medium",
	Low:    "low",
}

// ToString is a helper method to return the string of a ThresholdAlertSensitivity.
func (t ThresholdAlertSensitivity) ToString() string {
	return string(t)
}

// ToInt is a helper method to converts the ThresholdAlertSensitivity to its integer value.
func (t ThresholdAlertSensitivity) ToInt() int {
	switch t {
	case ThresholdAlertSensitives.High:
		return 1
	case ThresholdAlertSensitives.Medium:
		return 5
	case ThresholdAlertSensitives.Low:
		return 10
	default:
		return 5
	}
}
