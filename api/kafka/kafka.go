package kafka

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"time"
)

// Kafka represents Heroku's kafka APIs.
type Kafka struct {
	http   *simpleresty.Client
	config *config2.Config
}

// Response represents a generic Kafka response.
type Response struct {
	ID      *string `json:"id,omitempty"`
	Message *string `json:"message,omitempty"`
}

// New constructs a client to interface with the Heroku Kafka APIs.
func New(config *config2.Config) *Kafka {
	p := &Kafka{http: simpleresty.NewWithBaseURL(config.KafkaBaseURL), config: config}
	p.setHeaders()

	return p
}

func (k *Kafka) setHeaders() {
	k.http.SetHeader("Content-type", k.config.ContentTypeHeader).
		SetHeader("Accept", k.config.AcceptHeader).
		SetHeader("User-Agent", k.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Basic %s", k.config.BasicAuth)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if k.config.CustomHTTPHeaders != nil {
		k.http.SetHeaders(k.config.CustomHTTPHeaders)
	}
}
