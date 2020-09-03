package api

type Config struct {
	// MetricsBaseURL is the base URL for Heroku's metrics API.
	MetricsBaseURL string

	// PostgresBaseURL is the base URL for Heroku's postgres API.
	PostgresBaseURL string

	// UserAgent used when communicating with the Heroku API.
	UserAgent string

	// CustomHTTPHeaders are any additional user defined headers.
	CustomHTTPHeaders map[string]string

	// APIToken
	APIToken string

	// BasicAuth represents a base64 encoded string
	BasicAuth string

	// ContentTypeHeader
	ContentTypeHeader string

	// AcceptHeader
	AcceptHeader string
}

// parseOptions parses the supplied options functions.
func (c *Config) parseOptions(opts ...Option) error {
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
