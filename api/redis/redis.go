package redis

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"time"
)

// Redis represents Heroku's Redis APIs.
type Redis struct {
	http   *simpleresty.Client
	config *config2.Config
}

// New constructs a client to interface with the Heroku Redis APIs.
func New(config *config2.Config) *Redis {
	r := &Redis{http: simpleresty.NewWithBaseURL(config.PostgresBaseURL), config: config}
	r.setHeaders()

	return r
}

func (r *Redis) setHeaders() {
	r.http.SetHeader("Content-type", r.config.ContentTypeHeader).
		SetHeader("Accept", r.config.AcceptHeader).
		SetHeader("User-Agent", r.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", r.config.APIToken)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if r.config.CustomHTTPHeaders != nil {
		r.http.SetHeaders(r.config.CustomHTTPHeaders)
	}
}
