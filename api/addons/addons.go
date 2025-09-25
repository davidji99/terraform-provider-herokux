package addons

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"time"
)

// Addons represents Heroku's Addons APIs.
type Addons struct {
	http   *simpleresty.Client
	config *config2.Config
}

// New constructs a client to interface with the Heroku Addons APIs.
func New(config *config2.Config) *Addons {
	p := &Addons{http: simpleresty.NewWithBaseURL(config.KafkaBaseURL), config: config}
	p.setHeaders()

	return p
}

func (a *Addons) setHeaders() {
	a.http.SetHeader("Content-type", a.config.ContentTypeHeader).
		SetHeader("Accept", a.config.ContentTypeHeader).
		SetHeader("User-Agent", a.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", a.config.APIToken)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if a.config.CustomHTTPHeaders != nil {
		a.http.SetHeaders(a.config.CustomHTTPHeaders)
	}
}
