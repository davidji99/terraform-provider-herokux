package registry

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"time"
)

const (
	DistributionManifestAcceptHeader = "application/vnd.docker.distribution.manifest.v2+json"
)

// Registry represents API functionality that's part of the Heroku Registry.
type Registry struct {
	http   *simpleresty.Client
	config *config2.Config
}

// New constructs a client to interface with the Heroku Platform APIs.
func New(config *config2.Config) *Registry {
	p := &Registry{http: simpleresty.NewWithBaseURL(config.RegistryBaseURL), config: config}
	p.setHeaders()

	return p
}

func (r *Registry) setHeaders() {
	r.http.SetHeader("Content-type", r.config.ContentTypeHeader).
		SetHeader("Accept", DistributionManifestAcceptHeader).
		SetHeader("User-Agent", r.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", r.config.APIToken)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if r.config.CustomHTTPHeaders != nil {
		r.http.SetHeaders(r.config.CustomHTTPHeaders)
	}
}
