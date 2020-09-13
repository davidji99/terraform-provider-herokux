package data

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"time"
)

// Data represents Heroku's Data APIs.
type Data struct {
	http   *simpleresty.Client
	config *config2.Config
}

// New constructs a client to interface with the Heroku Data APIs.
func New(config *config2.Config) *Data {
	d := &Data{http: simpleresty.NewWithBaseURL(config.DataBaseURL), config: config}
	d.setHeaders()

	return d
}

func (d *Data) setHeaders() {
	d.http.SetHeader("Content-type", d.config.ContentTypeHeader).
		SetHeader("Accept", d.config.AcceptHeader).
		SetHeader("User-Agent", d.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", d.config.APIToken)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if d.config.CustomHTTPHeaders != nil {
		d.http.SetHeaders(d.config.CustomHTTPHeaders)
	}
}
