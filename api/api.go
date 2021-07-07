package api

import (
	"github.com/davidji99/terraform-provider-herokux/api/connect"
	"github.com/davidji99/terraform-provider-herokux/api/data"
	"github.com/davidji99/terraform-provider-herokux/api/kafka"
	"github.com/davidji99/terraform-provider-herokux/api/kolkrabbi"
	"github.com/davidji99/terraform-provider-herokux/api/metrics"
	config2 "github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/davidji99/terraform-provider-herokux/api/platform"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/davidji99/terraform-provider-herokux/api/redis"
	"github.com/davidji99/terraform-provider-herokux/api/registry"
	"github.com/davidji99/terraform-provider-herokux/api/scheduler"
)

const (
	// DefaultPlatformAPIBaseURL is the base Platform URL.
	DefaultPlatformAPIBaseURL = "https://api.heroku.com"

	// DefaultMetricAPIBaseURL is the default base Metric URL.
	DefaultMetricAPIBaseURL = "https://api.metrics.heroku.com"

	// DefaultPostgresAPIBaseURL is the default base URL for Postgres related APIs.
	DefaultPostgresAPIBaseURL = "https://postgres-api.heroku.com"

	// DefaultDataAPIBaseURL is the default base URL for the Data Graph APIs.
	DefaultDataAPIBaseURL = "https://data-api.heroku.com"

	// DefaultRedisAPIBaseURL is the default base URL for the Redis APIs.
	DefaultRedisAPIBaseURL = "https://redis-api.heroku.com"

	// DefaultConnectCentralBaseURL is the default base URL for Connect Central API.
	DefaultConnectCentralBaseURL = "https://hc-central.heroku.com"

	// DefaultConnectAPIBaseURL is the default base URL for the Connect APIs.
	// Setting to the 3-virginia endpoint by default.
	// Reference: https://devcenter.heroku.com/articles/heroku-connect-api#endpoints
	DefaultConnectAPIBaseURL = "https://connect-3-virginia.heroku.com/api/v3"

	// DefaultRegistryBaseURL is the default base URL for the Heroku Registry.
	DefaultRegistryBaseURL = "https://registry.heroku.com"

	// DefaultKolkrabbiAPIBaseURL is the default base URL for Kolkrabbi API.
	DefaultKolkrabbiAPIBaseURL = "https://kolkrabbi.heroku.com"

	// DefaultSchedulerAPIBaseURL is the default base URL for Scheduler API.
	DefaultSchedulerAPIBaseURL = "https://particleboard.heroku.com"

	// DefaultUserAgent is the user agent used when making API calls.
	DefaultUserAgent = "herokux-go"

	// DefaultAcceptHeader is the default Accept header.
	// TODO: see if this can be set back to just `application/json
	DefaultAcceptHeader = "application/vnd.heroku+json; version=3"

	// DefaultContentTypeHeader is the default and Content-Type header.
	DefaultContentTypeHeader = "application/json"
)

var (
	// DefaultMemberPermissions is a list of the default permissions for members/collaborators
	// on various Heroku resources.
	DefaultMemberPermissions = []string{"view", "deploy", "operate", "manage"}
)

// Client manages communication with various Heroku APIs.
type Client struct {
	config *config2.Config

	// API endpoints
	Connect   *connect.Connect
	Data      *data.Data
	Kafka     *kafka.Kafka
	Kolkrabbi *kolkrabbi.Kolkrabbi
	Metrics   *metrics.Metrics
	Platform  *platform.Platform
	Postgres  *postgres.Postgres
	Redis     *redis.Redis
	Registry  *registry.Registry
	Scheduler *scheduler.Scheduler
}

// New constructs a new client to interact with Heroku APIs.
func New(opts ...config2.Option) (*Client, error) {
	// Define baseline config values.
	config := &config2.Config{
		MetricsBaseURL:        DefaultMetricAPIBaseURL,
		PostgresBaseURL:       DefaultPostgresAPIBaseURL,
		KafkaBaseURL:          DefaultPostgresAPIBaseURL, // The Kafka API endpoints use the same base URL as postgres endpoints.
		DataBaseURL:           DefaultDataAPIBaseURL,
		PlatformBaseURL:       DefaultPlatformAPIBaseURL,
		RedisBaseURL:          DefaultRedisAPIBaseURL,
		ConnectBaseURL:        DefaultConnectAPIBaseURL,
		ConnectCentralBaseURL: DefaultConnectCentralBaseURL,
		RegistryBaseURL:       DefaultRegistryBaseURL,
		KolkrabbiURL:          DefaultKolkrabbiAPIBaseURL,
		SchedulerURL:          DefaultSchedulerAPIBaseURL,
		UserAgent:             DefaultUserAgent,
		APIToken:              "",
		BasicAuth:             "",
		ContentTypeHeader:     DefaultContentTypeHeader,
		AcceptHeader:          DefaultAcceptHeader,
	}

	// Define any user custom Client settings
	if optErr := config.ParseOptions(opts...); optErr != nil {
		return nil, optErr
	}

	// Construct new Client
	client := &Client{
		config:    config,
		Data:      data.New(config),
		Kafka:     kafka.New(config),
		Metrics:   metrics.New(config),
		Platform:  platform.New(config),
		Postgres:  postgres.New(config),
		Redis:     redis.New(config),
		Connect:   connect.New(config),
		Registry:  registry.New(config),
		Kolkrabbi: kolkrabbi.New(config),
		Scheduler: scheduler.New(config),
	}

	return client, nil
}
