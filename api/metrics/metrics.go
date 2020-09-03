package metrics

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"time"
)

// Metrics represents Heroku's Metrics APIs.
type Metrics struct {
	http   *simpleresty.Client
	config *config2.Config
}

// New constructs a client to interface with the Heroku Metrics APIs.
func New(config *config2.Config) *Metrics {
	m := &Metrics{http: simpleresty.NewWithBaseURL(config.MetricsBaseURL), config: config}
	m.setHeaders()

	return m
}

func (m *Metrics) setHeaders() {
	m.http.SetHeader("Content-type", m.config.ContentTypeHeader).
		SetHeader("Accept", m.config.AcceptHeader).
		SetHeader("User-Agent", m.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", m.config.APIToken)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if m.config.CustomHTTPHeaders != nil {
		m.http.SetHeaders(m.config.CustomHTTPHeaders)
	}
}
