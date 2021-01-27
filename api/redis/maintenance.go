package redis

import (
	"fmt"
	"regexp"

	"github.com/davidji99/simpleresty"
)

// MaintenanceWindowResponse represents the response from the Redis API when setting the maintenance window.
type MaintenanceWindowResponse struct {
	Window *string `json:"window,omitempty"`
}

// GetMaintenanceWindow returns the maintenance window for a redis database.
//
// All times are in UTC.
func (p *Redis) GetMaintenanceWindow(dbID string) (*GenericResponse, *simpleresty.Response, error) {
	var result *GenericResponse
	urlStr := p.http.RequestURL("/redis/v0/databases/%s/maintenance", dbID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// SetMaintenanceWindow sets the the weekly maintenance window for a redis database.
func (p *Redis) SetMaintenanceWindow(dbID, window string) (*MaintenanceWindowResponse, *simpleresty.Response, error) {
	var result *MaintenanceWindowResponse
	urlStr := p.http.RequestURL("/redis/v0/databases/%s/maintenance_window", dbID)

	// Validate the window string
	regex := regexp.MustCompile(`[A-Za-z]{2,10} \d\d?:[03]0$`)
	if !regex.MatchString(window) {
		return nil, nil, fmt.Errorf("window must be \"Day HH:MM\" where MM is 00 or 30")
	}

	body := struct {
		Description string `json:"description"`
	}{
		Description: window,
	}

	// Execute the request
	response, getErr := p.http.Put(urlStr, &result, &body)

	return result, response, getErr
}
