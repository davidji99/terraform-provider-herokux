package postgres

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"time"
)

// Postgres represents Heroku's postgres APIs.
type Postgres struct {
	http   *simpleresty.Client
	config *config2.Config
}

// GenericResponse represents a generic response from the Postgres API.
type GenericResponse struct {
	ID      *string `json:"id,omitempty"`
	Message *string `json:"message,omitempty"`
}

// New constructs a client to interface with the Heroku Postgres APIs.
func New(config *config2.Config) *Postgres {
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
