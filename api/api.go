package api

import (
	"github.com/davidji99/terraform-provider-herokux/api/metrics"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
)

const (
	// DefaultAPIBaseURL is the base URL.
	DefaultAPIBaseURL = "https://api.heroku.com"

	// DefaultMetricAPIBaseURL is the base Metric URL.
	DefaultMetricAPIBaseURL = "https://api.metrics.heroku.com"

	// DefaultMetricAPIBaseURL
	DefaultPostgresAPIBaseURL = "https://postgres-api.heroku.com/postgres/v0"

	// DefaultUserAgent is the user agent used when making API calls.
	DefaultUserAgent = "herokux-go"

	// DefaultAcceptHeader is the default and required Accept header.
	DefaultAcceptHeader = "application/vnd.heroku+json; version=3"

	// DefaultContentTypeHeader
	DefaultContentTypeHeader = "application/json"
)

// A Client manages communication with the Heroku API.
type Client struct {
	config *Config

	// API endpoints
	Metrics  *metrics.Metrics
	Postgres *postgres.Postgres
}

// New constructs a new client to interact with the API.
func New(opts ...Option) (*Client, error) {
	// Define baseline config values.
	config := &Config{
		MetricsBaseURL:    DefaultMetricAPIBaseURL,
		PostgresBaseURL:   DefaultPostgresAPIBaseURL,
		UserAgent:         DefaultUserAgent,
		APIToken:          "",
		BasicAuth:         "",
		ContentTypeHeader: DefaultContentTypeHeader,
		AcceptHeader:      DefaultAcceptHeader,
	}

	// Define any user custom Client settings
	if optErr := config.parseOptions(opts...); optErr != nil {
		return nil, optErr
	}

	// Construct new Client
	client := &Client{
		config:   config,
		Metrics:  metrics.New(config),
		Postgres: postgres.New(config),
	}

	return client, nil
}
