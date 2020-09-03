package metrics

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	"github.com/davidji99/terraform-provider-herokux/api"
	"time"
)

type Metrics struct {
	http   *simpleresty.Client
	config *api.Config
}

func New(config *api.Config) *Metrics {
	m := &Metrics{http: simpleresty.NewWithBaseURL(config.MetricsBaseURL), config: config}
	m.setHeaders()

	return m
}

func (m *Metrics) setHeaders() {
	m.http.SetHeader("Content-type", m.config.ContentTypeHeader).
		SetHeader("Accept", m.config.AcceptHeader).
		SetHeader("User-Agent", m.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Basic %s", m.config.BasicAuth)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if m.config.CustomHTTPHeaders != nil {
		m.http.SetHeaders(m.config.CustomHTTPHeaders)
	}
}
