package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/davidji99/simpleresty"
	"log"
)

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

// ListMonitors lists all monitors for a formation.
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

// FindMonitorByName gets a single monitor for a formation by its associated app ID and formation name/process type.
func (m *Metrics) FindMonitorByName(appID, formationName string) (*FormationMonitor, *simpleresty.Response, error) {
	monitors, response, listErr := m.ListMonitors(appID, formationName)
	if listErr != nil {
		return nil, response, listErr
	}

	for _, m := range monitors {
		if m.GetAppID() == appID && m.GetProcessType() == formationName {
			return m, nil, nil
		}
	}

	return nil, nil, fmt.Errorf("did not find a monitor for app %s's formation %s", appID, formationName)
}

// SetAutoscale modifies the autoscaling properties for an app dyno formation.
func (m *Metrics) SetAutoscale(appID, formationName, monitorID string, opts *AutoscalingRequest) (bool, *simpleresty.Response, error) {
	urlStr := m.http.RequestURL("/apps/%s/formation/%s/monitors/%s", appID, formationName, monitorID)

	opts.Operation = "GREATER_OR_EQUAL"

	log.Printf("%+v\n", opts)

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
