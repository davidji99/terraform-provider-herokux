package postgres

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	"regexp"
)

type MaintenanceWindowResponse struct {
	Window *string `json:"window,omitempty"`
}

// GetMaintenanceWindow returns the maintenance window for a postgres database.
//
// All times are in UTC.
func (p *Postgres) GetMaintenanceWindow(dbID string) (*GenericResponse, *simpleresty.Response, error) {
	var result *GenericResponse
	urlStr := p.http.RequestURL("/client/v11/databases/%s/maintenance", dbID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// SetMaintenanceWindow sets the the weekly maintenance window for a postgres database.
func (p *Postgres) SetMaintenanceWindow(dbID, window string) (*MaintenanceWindowResponse, *simpleresty.Response, error) {
	var result *MaintenanceWindowResponse
	urlStr := p.http.RequestURL("/client/v11/databases/%s/maintenance_window", dbID)

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
