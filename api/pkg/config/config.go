package config

// Config represents all configuration options available to user to customize the Client.
type Config struct {
	// MetricsBaseURL is the base URL for Heroku's metrics API.
	MetricsBaseURL string

	// PostgresBaseURL is the base URL for Heroku's postgres APIs.
	PostgresBaseURL string

	// KafkaBaseURL is the base URL for Heroku's kafka APIs.
	KafkaBaseURL string

	// DataBaseURL is the base URL for Heroku's Data APIs.
	DataBaseURL string

	// PlatformBaseURL is the base URL for Heroku's Platform APIs.
	PlatformBaseURL string

	// RedisBaseURL is the base URL for Heroku's Redis APIs.
	RedisBaseURL string

	// ConnectBaseURL is the base URL for Heroku's Connect APIs.
	ConnectBaseURL string

	// RegistryBaseURL is the base URL for Heroku's Registry.
	RegistryBaseURL string

	// UserAgent used when communicating with the Heroku API.
	UserAgent string

	// CustomHTTPHeaders are any additional user defined headers.
	CustomHTTPHeaders map[string]string

	// APIToken is the Heroku API key.
	APIToken string

	// BasicAuth represents a base64 encoded string
	BasicAuth string

	// ContentTypeHeader
	ContentTypeHeader string

	// AcceptHeader
	AcceptHeader string
}

// parseOptions parses the supplied options functions.
func (c *Config) ParseOptions(opts ...Option) error {
	// Range over each options function and apply it to our API type to
	// configure it. Options functions are applied in order, with any
	// conflicting options overriding earlier calls.
	for _, option := range opts {
		err := option(c)
		if err != nil {
			return err
		}
	}

	return nil
}
