package config

import (
	"encoding/base64"
	"fmt"
)

// Option is a functional option for configuring the API client.
type Option func(*Config) error

// MetricsBaseURL allows for a custom Metrics API URL.
func MetricsBaseURL(url string) Option {
	return func(c *Config) error {
		if err := validateBaseURLOption(url); err != nil {
			return err
		}

		c.MetricsBaseURL = url
		return nil
	}
}

// PostgresBaseURL allows for a custom Postgres API URL.
func PostgresBaseURL(url string) Option {
	return func(c *Config) error {
		if err := validateBaseURLOption(url); err != nil {
			return err
		}

		c.PostgresBaseURL = url
		return nil
	}
}

// KafkaBaseURL allows for a custom Kafka API URL.
func KafkaBaseURL(url string) Option {
	return func(c *Config) error {
		if err := validateBaseURLOption(url); err != nil {
			return err
		}

		c.KafkaBaseURL = url
		return nil
	}
}

// DataBaseURL allows for a custom Data API URL.
func DataBaseURL(url string) Option {
	return func(c *Config) error {
		if err := validateBaseURLOption(url); err != nil {
			return err
		}

		c.DataBaseURL = url
		return nil
	}
}

// PlatformBaseURL allows for a custom Platform API URL.
func PlatformBaseURL(url string) Option {
	return func(c *Config) error {
		if err := validateBaseURLOption(url); err != nil {
			return err
		}

		c.PlatformBaseURL = url
		return nil
	}
}

// UserAgent allows for a custom User Agent.
func UserAgent(userAgent string) Option {
	return func(c *Config) error {
		c.UserAgent = userAgent
		return nil
	}
}

// CustomHTTPHeaders allows for additional HTTPHeaders.
func CustomHTTPHeaders(headers map[string]string) Option {
	return func(c *Config) error {
		c.CustomHTTPHeaders = headers
		return nil
	}
}

// APIToken sets the API token for authentication.
func APIToken(token string) Option {
	return func(c *Config) error {
		c.APIToken = token
		return nil
	}
}

// BasicAuth sets Base64 encoding of the username and password parameters joined by a single colon (:).
func BasicAuth(username, password string) Option {
	return func(c *Config) error {
		userPass := fmt.Sprintf("%s:%s", username, password)
		c.BasicAuth = base64.StdEncoding.EncodeToString([]byte(userPass))

		return nil
	}
}

// ContentTypeHeader allows for a custom Content-Type header.
func ContentTypeHeader(s string) Option {
	return func(c *Config) error {
		c.ContentTypeHeader = s
		return nil
	}
}

// AcceptHeader allows for a custom Aceept header.
func AcceptHeader(s string) Option {
	return func(c *Config) error {
		c.AcceptHeader = s
		return nil
	}
}

// validateBaseURLOption ensures that any custom base URLs do not end with a trailing slash.
func validateBaseURLOption(url string) error {
	// Validate that there is no trailing slashes before setting the custom baseURL
	if url[len(url)-1:] == "/" {
		return fmt.Errorf("custom base URL cannot contain a trailing slash")
	}
	return nil
}
