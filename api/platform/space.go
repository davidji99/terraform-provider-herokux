package platform

import (
	"fmt"

	"github.com/davidji99/simpleresty"
)

// GetSpaceLogDrain returns a space's log drain if available.
func (p *Platform) GetSpaceLogDrain(spaceID string) (*LogDrain, *simpleresty.Response, error) {
	var result *LogDrain
	urlStr := p.space.RequestURL("/spaces/%s/log-drain", spaceID)

	// Execute the request
	response, getErr := p.space.Get(urlStr, &result, nil)

	if getErr != nil && response != nil {
		return nil, response, fmt.Errorf("space log drain: HTTP %d", response.StatusCode)
	}
	return result, response, getErr
}

// SetSpaceLogDrain sets a space's log drain.
//
// To remove a log drain, pass in an empty string.
func (p *Platform) SetSpaceLogDrain(spaceID string, url string) (*LogDrain, *simpleresty.Response, error) {
	var result *LogDrain
	urlStr := p.space.RequestURL("/spaces/%s/log-drain", spaceID)

	opts := struct {
		Url string `json:"url"`
	}{
		Url: url,
	}

	// Execute the request
	response, getErr := p.space.Put(urlStr, &result, opts)

	if getErr != nil && response != nil {
		return nil, response, fmt.Errorf("space log drain: HTTP %d", response.StatusCode)
	}
	return result, response, getErr
}
