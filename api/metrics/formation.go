package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/davidji99/simpleresty"
)

const (
	// DefaultOperationAttrVal defines the only possible value for `op` or operation.
	DefaultOperationAttrVal = "GREATER_OR_EQUAL"
)

var (
	// AlertSensitivityValues represents the possible numerical values
	// for an alert's sensitivity.
	AlertSensitivityValues = []int{1, 5, 10}

	// AlertReminderFrequencies represents the possible numerical values
	// for an alert's reminder frequency.
	AlertReminderFrequencies = []int{5, 60, 1440}
)

// FormationMonitor represents a formation monitor. A monitor can be an app's autoscaling settings
// or an app alert.
type FormationMonitor struct {
	// ID is the UUID of the monitor.
	ID *string `json:"id,omitempty"`

	// AppID is the app UUID.
	AppID *string `json:"app_id,omitempty"`

	// MetricUUID
	MetricUUID *string `json:"metric_uuid,omitempty"`

	// ProcessType is the name of the dyno such as "web".
	ProcessType *string `json:"process_type,omitempty"`

	// Name represents the name of the monitor.
	// Possible values are "LATENCY", "LATENCY_SCALE", and "ERROR_RATE"
	Name *FormationMonitorName `json:"name,omitempty"`

	// Operation. The only currently supported value is "GREATER_OR_EQUAL".
	Operation *string `json:"op,omitempty"`

	// Period represents the alert's sensitivity in minutes and is not used for autoscaling.
	// Possible values are:
	// - 1 "high"
	// - 5 "medium"
	// - 10 "low"
	Period *int `json:"period,omitempty"`

	// IsActive determines whether or not the monitor is on or off.
	IsActive *bool `json:"is_active"`

	// State represents the state of the monitor.
	State *string `json:"state,omitempty"`

	// ActionType represents the formation monitor's action type.
	// There are only two possible values here: "alert" and "scale".
	ActionType *FormationMonitorActionType `json:"action_type,omitempty"`

	// NotificationChannels represents destinations to send notifications.
	// In most cases, enabling notifications results in this field being set to `["app"]`.
	// For app alerts, additional (confirmed) email addresses are set here as well.
	NotificationChannels []string `json:"notification_channels,omitempty"`

	// NotificationPeriod represents the frequency of alert reminders.
	// The possible values returned here are 5, 60, or 1440 minutes.
	NotificationPeriod *int `json:"notification_period"`

	// MinQuantity represents the minimum number of dynos for autoscaling.
	MinQuantity *int `json:"min_quantity"`

	// MaxQuantity represents the maximum number of dynos for autoscaling.
	MaxQuantity *int `json:"max_quantity"`

	// ForecastPeriod
	ForecastPeriod *int `json:"forecast_period,omitempty"`

	// Value represent the alert threshold or an autoscaling's desired p95 response time.
	Value *json.Number `json:"value,omitempty"`
}

// FormationAutoscalingRequest represents a request to modify autoscaling for an app formation.
type FormationAutoscalingRequest struct {
	IsActive             bool                       `json:"is_active"`
	Quantity             int                        `json:"quantity"`
	MaxQuantity          int                        `json:"max_quantity,omitempty"`
	MinQuantity          int                        `json:"min_quantity,omitempty"`
	NotificationPeriod   int                        `json:"notification_period"`
	DesiredP95RespTime   int                        `json:"value,omitempty"`
	Period               int                        `json:"period,omitempty"`
	DynoSize             string                     `json:"dyno_size,omitempty"`
	ActionType           FormationMonitorActionType `json:"action_type,omitempty"`
	Operation            string                     `json:"op,omitempty"`
	Name                 FormationMonitorName       `json:"name,omitempty"`
	NotificationChannels []string                   `json:"notification_channels"`
}

// FormationAlertRequest represents a request to modify an alert for an app formation.
type FormationAlertRequest struct {
	IsActive             bool                       `json:"is_active"`
	ReminderFrequency    int                        `json:"notification_period"`
	Sensitivity          int                        `json:"period,omitempty"`
	Threshold            json.Number                `json:"value,omitempty"`
	DynoSize             string                     `json:"dyno_size,omitempty"`
	Operation            string                     `json:"op,omitempty"`
	NotificationChannels []string                   `json:"notification_channels"`
	ActionType           FormationMonitorActionType `json:"action_type,omitempty"`
	Name                 FormationMonitorName       `json:"name,omitempty"`
}

// ListMonitors lists all monitors for a formation.
//
// This endpoint returns 200 and an empty array even if the app has no dyno and/or process type ('web')  associated to it.
func (m *Metrics) ListMonitors(appID, formationName string) ([]*FormationMonitor, *simpleresty.Response, error) {
	var result []*FormationMonitor
	urlStr := m.http.RequestURL("/apps/%s/formation/%s/monitors", appID, formationName)

	// Execute the request
	response, getErr := m.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// GetMonitor gets a single monitor for a formation.
//
// This endpoint returns text/plain; charset=utf-8 despite passing in the correct request headers.
// This we need to manually unmarshall the response into the appropriate struct.
//
// This method also is used to return app threshold alerts.
func (m *Metrics) GetMonitor(appID, formationName, monitorID string) (*FormationMonitor, *simpleresty.Response, error) {
	var result FormationMonitor
	urlStr := m.http.RequestURL("/apps/%s/formation/%s/monitors/%s", appID, formationName, monitorID)

	// Execute the request
	response, getErr := m.http.Get(urlStr, nil, nil)
	if getErr != nil {
		return nil, response, getErr
	}

	unmarshallErr := json.Unmarshal([]byte(response.Body), &result)
	if unmarshallErr != nil {
		return nil, response, unmarshallErr
	}

	return &result, response, nil
}

// DeleteMonitor destroys an existing formation monitor.
//
// Returns '202 Accepted' if successful. WARNING! This method may cause unknown issues if used.
func (m *Metrics) DeleteMonitor(appID, formationName, monitorID string) (*simpleresty.Response, error) {
	urlStr := m.http.RequestURL("/apps/%s/formation/%s/monitors/%s", appID, formationName, monitorID)

	// Execute the request
	response, deleteErr := m.http.Delete(urlStr, nil, nil)

	return response, deleteErr
}

// FindMonitorByName gets a single monitor for a formation by its name.
func (m *Metrics) FindMonitorByName(appID, formationName string, name FormationMonitorName) (*FormationMonitor, *simpleresty.Response, error) {
	monitors, response, listErr := m.ListMonitors(appID, formationName)
	if listErr != nil {
		return nil, response, listErr
	}

	for _, m := range monitors {
		if m.GetAppID() == appID && m.GetProcessType() == formationName && *m.GetName() == name {
			return m, nil, nil
		}
	}

	return nil, nil, fmt.Errorf("did not find a %s monitor for app %s's formation %s",
		name.ToString(), appID, formationName)
}

// CreateFormationAutoscaling sets up autoscaling for an app formation.
//
// The API response body only has the formation autoscaling UUID.
func (m *Metrics) CreateFormationAutoscaling(appID, formationName string, opts *FormationAutoscalingRequest) (*FormationMonitor, *simpleresty.Response, error) {
	var result *FormationMonitor

	urlStr := m.http.RequestURL("/apps/%s/formation/%s/monitors", appID, formationName)

	// Execute the request
	response, createErr := m.http.Post(urlStr, &result, opts)

	return result, response, createErr
}

// UpdateFormationAutoscaling modifies autoscaling for an app formation.
//
// The endpoint does not return any response. Instead, the method returns true if request is successful; false otherwise,
func (m *Metrics) UpdateFormationAutoscaling(appID, formationName, monitorID string, opts *FormationAutoscalingRequest) (bool, *simpleresty.Response, error) {
	urlStr := m.http.RequestURL("/apps/%s/formation/%s/monitors/%s", appID, formationName, monitorID)

	// Execute the request
	response, updateErr := m.http.Patch(urlStr, nil, opts)
	if updateErr != nil {
		return false, response, updateErr
	}

	if response.StatusCode == 202 {
		return true, response, nil
	}

	return false, response, fmt.Errorf("did not properly update %s's %s formation autoscaling", appID, formationName)
}

// CreateFormationAlert creates an alert for an app formation.
//
// The API response body only has the formation alert UUID.
func (m *Metrics) CreateFormationAlert(appID, formationName string, opts *FormationAlertRequest) (*FormationMonitor, *simpleresty.Response, error) {
	var result *FormationMonitor

	urlStr := m.http.RequestURL("/apps/%s/formation/%s/monitors", appID, formationName)

	// Execute the request
	response, createErr := m.http.Post(urlStr, &result, opts)

	return result, response, createErr
}

// UpdateFormationAlert updates an existing alert for an app formation.
func (m *Metrics) UpdateFormationAlert(appID, formationName, alertID string, opts *FormationAlertRequest) (bool, *simpleresty.Response, error) {
	var result *FormationMonitor

	urlStr := m.http.RequestURL("/apps/%s/formation/%s/monitors/%s", appID, formationName, alertID)

	// Execute the request
	response, updateErr := m.http.Patch(urlStr, &result, opts)
	if updateErr != nil {
		return false, response, updateErr
	}

	if response.StatusCode == 202 {
		return true, response, nil
	}

	return false, response, fmt.Errorf("did not properly update %s's %s alert", appID, formationName)
}
