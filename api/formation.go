package api

// FormationMonitor represents a formation monitor.
type FormationMonitor struct {
	ID                   *string  `json:"id,omitempty"`
	AppID                *string  `json:"app_id,omitempty"`
	MetricUUID           *string  `json:"metric_uuid,omitempty"`
	ProcessType          *string  `json:"process_type,omitempty"`
	Name                 *string  `json:"name,omitempty"`
	Value                *int     `json:"value,omitempty"`
	Operation            *string  `json:"op,omitempty"`
	Period               *int     `json:"period,omitempty"`
	IsActive             *bool    `json:"is_active"`
	State                *string  `json:"state,omitempty"`
	ActionType           *string  `json:"action_type,omitempty"`
	NotificationChannels []string `json:"notification_channels,omitempty"`
	NotificationPeriod   *int     `json:"notification_period"`
	MinQuantity          *int     `json:"min_quantity"`
	MaxQuantity          *int     `json:"max_quantity"`
	ForecastPeriod       *int     `json:"forecast_period,omitempty"`
}

// AutoscalingRequest represents a request to autoscale an app dyno's formation.
type AutoscalingRequest struct {
	DynoSize             string   `json:"dyno_size,omitempty"`
	IsActive             bool     `json:"is_active"`
	MaxQuantity          int      `json:"max_quantity,omitempty"`
	MinQuantity          int      `json:"min_quantity,omitempty"`
	NotificationChannels []string `json:"notification_channels"`
	NotificationPeriod   int      `json:"notification_period"`
	DesiredP95RespTime   int      `json:"value,omitempty"`
	Period               int      `json:"period,omitempty"`
	ActionType           string   `json:"action_type,omitempty"`
	Operation            string   `json:"op,omitempty"`
	Quantity             int      `json:"quantity,omitempty"`
}
