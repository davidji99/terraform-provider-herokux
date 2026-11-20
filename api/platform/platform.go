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
	DogwoodAcceptHeader               = "application/vnd.heroku+json; version=3.dogwood"
	WebhooksAcceptHeader              = "application/vnd.heroku+json; version=3.webhooks"
	DockerReleasesAcceptHeader        = "application/vnd.heroku+json; version=3.docker-releases"
	PipelineCollaboratorsAcceptHeader = "application/vnd.heroku+json; version=3.pipeline-collaborators"
)

// Platform represents Heroku's Platform API.
//
// The APIs under Platform almost exclusively consist of variants to the official Platform API.
type Platform struct {
	http      *simpleresty.Client
	webhooks  *simpleresty.Client
	teamOAuth *simpleresty.Client
	space     *simpleresty.Client
	pipelines *simpleresty.Client
	formation *simpleresty.Client
}

// New constructs a client to interface with the Heroku Platform APIs.
func New(config *config2.Config) *Platform {
	return &Platform{
		http:      newPlatformClient(config, DefaultAcceptHeader),
		webhooks:  newPlatformClient(config, WebhooksAcceptHeader),
		teamOAuth: newPlatformClient(config, "application/vnd.heroku+json; version=3.sdk"),
		space:     newPlatformClient(config, DogwoodAcceptHeader),
		pipelines: newPlatformClient(config, PipelineCollaboratorsAcceptHeader),
		formation: newPlatformClient(config, DockerReleasesAcceptHeader),
	}
}

func newPlatformClient(config *config2.Config, accept string) *simpleresty.Client {
	client := simpleresty.NewWithBaseURL(config.PlatformBaseURL)
	client.SetHeader("Content-type", config.ContentTypeHeader).
		SetHeader("User-Agent", config.UserAgent).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.APIToken)).
		SetTimeout(2 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	client.SetHeaders(config.CustomHTTPHeaders)
	client.SetHeader("Accept", accept)
	return client
}
