package api

import (
	"github.com/davidji99/terraform-provider-herokux/api/data"
	"github.com/davidji99/terraform-provider-herokux/api/kafka"
	"github.com/davidji99/terraform-provider-herokux/api/metrics"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
)

const (
	// DefaultAPIBaseURL is the base URL.
	DefaultAPIBaseURL = "https://api.heroku.com"

	// DefaultMetricAPIBaseURL is the default base Metric URL.
	DefaultMetricAPIBaseURL = "https://api.metrics.heroku.com"

	// DefaultPostgresAPIBaseURL is the default base URL for Postgres related APIs.
	DefaultPostgresAPIBaseURL = "https://postgres-api.heroku.com"

	// DefaultKafkaAPIBaseURL is the default base URL for Kafka related APIs.
	DefaultKafkaAPIBaseURL = "https://postgres-api.heroku.com/data/kafka/v0"

	// DefaultDataBaseURL is the default base URL for the Data Graph APIs.
	DefaultDataBaseURL = "https://data-api.heroku.com"

	// DefaultUserAgent is the user agent used when making API calls.
	DefaultUserAgent = "herokux-go"

	// DefaultAcceptHeader is the default Accept header.
	DefaultAcceptHeader = "application/vnd.heroku+json; version=3"

	// DefaultContentTypeHeader is the default and Content-Type header.
	DefaultContentTypeHeader = "application/json"
)

// Client manages communication with various Heroku APIs.
type Client struct {
	config *config2.Config

	// API endpoints
	Data     *data.Data
	Kafka    *kafka.Kafka
	Metrics  *metrics.Metrics
	Postgres *postgres.Postgres
}

// New constructs a new client to interact with Heroku APIs.
func New(opts ...config2.Option) (*Client, error) {
	// Define baseline config values.
	config := &config2.Config{
		MetricsBaseURL:    DefaultMetricAPIBaseURL,
		PostgresBaseURL:   DefaultPostgresAPIBaseURL,
		KafkaBaseURL:      DefaultKafkaAPIBaseURL,
		DataBaseURL:       DefaultDataBaseURL,
		UserAgent:         DefaultUserAgent,
		APIToken:          "",
		BasicAuth:         "",
		ContentTypeHeader: DefaultContentTypeHeader,
		AcceptHeader:      DefaultAcceptHeader,
	}

	// Define any user custom Client settings
	if optErr := config.ParseOptions(opts...); optErr != nil {
		return nil, optErr
	}

	// Construct new Client
	client := &Client{
		config:   config,
		Data:     data.New(config),
		Kafka:    kafka.New(config),
		Metrics:  metrics.New(config),
		Postgres: postgres.New(config),
	}

	return client, nil
}
