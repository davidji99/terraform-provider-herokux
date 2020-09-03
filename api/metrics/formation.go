package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/davidji99/simpleresty"
	"github.com/davidji99/terraform-provider-herokux/api"
	"log"
)

// ListMonitors lists all monitors for a formation.
func (m *Metrics) ListMonitors(appID, formationName string) ([]*api.FormationMonitor, *simpleresty.Response, error) {
	var result []*api.FormationMonitor
	urlStr := m.http.RequestURL("/apps/%s/formation/%s/monitors", appID, formationName)

	// Execute the request
	response, getErr := m.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// GetMonitor gets a single monitor for a formation.
//
// This endpoint returns text/plain; charset=utf-8 despite passing in the correct request headers.
// This we need to manually unmarshall the response into the appropriate struct.
func (m *Metrics) GetMonitor(appID, formationName, monitorID string) (*api.FormationMonitor, *simpleresty.Response, error) {
	var result api.FormationMonitor
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
func (m *Metrics) FindMonitorByName(appID, formationName string) (*api.FormationMonitor, *simpleresty.Response, error) {
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
func (m *Metrics) SetAutoscale(appID, formationName, monitorID string, opts *api.AutoscalingRequest) (bool, *simpleresty.Response, error) {
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
