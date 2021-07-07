package scheduler

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	config "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"time"
)

// Scheduler represents API functionality for the Heroku scheduler addon.
type Scheduler struct {
	http   *simpleresty.Client
	config *config.Config
}

// New constructs a client to interface with the Heroku scheduler APIs.
func New(config *config.Config) *Scheduler {
	s := &Scheduler{http: simpleresty.NewWithBaseURL(config.SchedulerURL), config: config}
	s.setHeaders()

	return s
}

func (s *Scheduler) setHeaders() {
	s.http.SetHeader("Content-type", "application/vnd.api+json").
		SetHeader("Accept", "application/vnd.api+json").
		SetHeader("User-Agent", s.config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", s.config.APIToken)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if s.config.CustomHTTPHeaders != nil {
		s.http.SetHeaders(s.config.CustomHTTPHeaders)
	}
}
