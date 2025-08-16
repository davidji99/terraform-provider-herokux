package kolkrabbi

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"time"
)

// Kolkrabbi represents Heroku's kolkrabbi APIs.
type Kolkrabbi struct {
	http   *simpleresty.Client
	config *config2.Config
}

// New constructs a client to interface with the Heroku Postgres APIs.
func New(config *config2.Config) *Kolkrabbi {
	k := &Kolkrabbi{http: simpleresty.NewWithBaseURL(config.KolkrabbiBaseURL), config: config}
	k.setHeaders()

	return k
}

func (k *Kolkrabbi) setHeaders() {
	k.http.SetHeader("Content-type", k.config.ContentTypeHeader).
		SetHeader("Accept", k.config.AcceptHeader).
		SetHeader("User-Agent", k.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", k.config.APIToken)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if k.config.CustomHTTPHeaders != nil {
		k.http.SetHeaders(k.config.CustomHTTPHeaders)
	}
}
