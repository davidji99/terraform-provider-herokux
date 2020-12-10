package connect

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"time"
)

const (
	ConnectCentralBaseAPIURL = "https://hc-central.heroku.com"
)

// Connect represents Heroku's Connect APIs.
type Connect struct {
	http   *simpleresty.Client
	config *config2.Config
}

// New constructs a client to interface with the Heroku Connect APIs.
func New(config *config2.Config) *Connect {
	c := &Connect{http: simpleresty.New(), config: config}
	c.setHeaders()
	return c
}

// SetRootAPIBaseURL determines and sets the client's base URL based on the specified appID & connectID arguments.
//
// This is required.
func (c *Connect) SetRootAPIBaseURL(appID, connectID string) error {
	c.setHeaders()
	c.http.SetBaseURL(ConnectCentralBaseAPIURL)

	var result *AuthResponse
	urlStr := c.http.RequestURL("/auth/%s", appID)

	// Execute the request
	_, postErr := c.http.Post(urlStr, &result, nil)
	if postErr != nil {
		return postErr
	}

	// Loop through all connections and find the specified one by its ID.
	for _, connection := range result.Connections {
		if connection.GetID() == connectID {
			// Then when the target connection is found, set the client's base URL to the region URL.
			c.http.SetBaseURL(connection.GetRegionURL())
			return nil
		}
	}

	return fmt.Errorf("did not find connection %s on app %s", connectID, appID)
}

func (c *Connect) setHeaders() {
	c.http.SetHeader("Content-type", c.config.ContentTypeHeader).
		SetHeader("Accept", c.config.ContentTypeHeader).
		SetHeader("User-Agent", c.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.config.APIToken)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if c.config.CustomHTTPHeaders != nil {
		c.http.SetHeaders(c.config.CustomHTTPHeaders)
	}
}
