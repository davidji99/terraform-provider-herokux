package platform

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"time"
)

const (
	// DefaultAcceptHeader is the default Accept header.
	DefaultAcceptHeader = "application/vnd.heroku+json; version=3"

	// The following are various Accept headers for Platform API variants.
	DogwoodAcceptHeader  = "application/vnd.heroku+json; version=3.dogwood"
	WebhooksAcceptHeader = "application/vnd.heroku+json; version=3.webhooks"
)

// Platform represents Heroku's Platform API.
//
// The APIs under Platform almost exclusively consist of variants to the official Platform API.
type Platform struct {
	http   *simpleresty.Client
	config *config2.Config
}

// New constructs a client to interface with the Heroku Platform APIs.
func New(config *config2.Config) *Platform {
	p := &Platform{http: simpleresty.NewWithBaseURL(config.PlatformBaseURL), config: config}
	p.setHeaders()

	return p
}

func (p *Platform) setHeaders() {
	p.http.SetHeader("Content-type", p.config.ContentTypeHeader).
		SetHeader("Accept", DefaultAcceptHeader).
		SetHeader("User-Agent", p.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", p.config.APIToken)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if p.config.CustomHTTPHeaders != nil {
		p.http.SetHeaders(p.config.CustomHTTPHeaders)
	}
}

// resetAcceptHeader sets the `Accept` header back to the default value.
func (p *Platform) resetAcceptHeader() {
	p.http.SetHeader("Accept", DefaultAcceptHeader)
}
