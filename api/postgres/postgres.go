package postgres

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	"github.com/davidji99/terraform-provider-herokux/api"
	"time"
)

type Postgres struct {
	http   *simpleresty.Client
	config *api.Config
}

func New(config *api.Config) *Postgres {
	p := &Postgres{http: simpleresty.NewWithBaseURL(config.PostgresBaseURL), config: config}
	p.setHeaders()

	return p
}

func (p *Postgres) setHeaders() {
	p.http.SetHeader("Content-type", p.config.ContentTypeHeader).
		SetHeader("Accept", p.config.AcceptHeader).
		SetHeader("User-Agent", p.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Basic %s", p.config.BasicAuth)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if p.config.CustomHTTPHeaders != nil {
		p.http.SetHeaders(p.config.CustomHTTPHeaders)
	}
}
