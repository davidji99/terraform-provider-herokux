package platform

import (
	"github.com/davidji99/simpleresty"
)

// App represents a Heroku app.
type App struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// AddonResponse represents a generic response from the Platform Addon API.
type AddonResponse struct {
	ID           *string `json:"id,omitempty"`
	Name         *string `json:"name,omitempty"`
	App          *App    `json:"app,omitempty"`
	AddonService *App    `json:"addon_service,omitempty"`
}

// ListAppAddons returns the maintenance window for a postgres database.
//
// All times are in UTC.
func (p *Platform) ListAppAddons(appID string) ([]*AddonResponse, *simpleresty.Response, error) {
	var result []*AddonResponse
	urlStr := p.http.RequestURL("/apps/%s/addons", appID)
	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}
