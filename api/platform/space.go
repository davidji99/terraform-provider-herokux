package platform

import "github.com/davidji99/simpleresty"

// GetSpaceLogDrain returns a space's log drain if available.
func (p *Platform) GetSpaceLogDrain(spaceID string) (*LogDrain, *simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	var result *LogDrain
	urlStr := p.http.RequestURL("/spaces/%s/log-drain", spaceID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", DogwoodAcceptHeader)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// SetSpaceLogDrain sets a space's log drain.
//
// To remove a log drain, pass in an empty string.
func (p *Platform) SetSpaceLogDrain(spaceID string, url string) (*LogDrain, *simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	var result *LogDrain
	urlStr := p.http.RequestURL("/spaces/%s/log-drain", spaceID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", DogwoodAcceptHeader)

	opts := struct {
		Url string `json:"url"`
	}{
		Url: url,
	}

	// Execute the request
	response, getErr := p.http.Put(urlStr, &result, opts)

	return result, response, getErr
}
