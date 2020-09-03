package config

import (
	"encoding/base64"
	"fmt"
)

// Option is a functional option for configuring the API client.
type Option func(*Config) error

// MetricsBaseURL
func MetricsBaseURL(url string) Option {
	return func(c *Config) error {
		if err := validateBaseURLOption(url); err != nil {
			return err
		}

		c.MetricsBaseURL = url
		return nil
	}
}

// PostgresBaseURL
func PostgresBaseURL(url string) Option {
	return func(c *Config) error {
		if err := validateBaseURLOption(url); err != nil {
			return err
		}

		c.PostgresBaseURL = url
		return nil
	}
}

// UserAgent allows overriding of the default User Agent.
func UserAgent(userAgent string) Option {
	return func(c *Config) error {
		c.UserAgent = userAgent
		return nil
	}
}

// CustomHTTPHeaders sets additional HTTPHeaders
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

// BasicAuth takes an username and password parameter & sets Base64 encoding of the two parameters joined by a single colon (:).
func BasicAuth(username, password string) Option {
	return func(c *Config) error {
		if username == "" || password == "" {
			return fmt.Errorf("both username and password must be set for Basic authentication")
		}

		userPass := fmt.Sprintf("%s:%s", username, password)
		c.BasicAuth = base64.StdEncoding.EncodeToString([]byte(userPass))

		return nil
	}
}

// ContentTypeHeader
func ContentTypeHeader(s string) Option {
	return func(c *Config) error {
		c.ContentTypeHeader = s
		return nil
	}
}

// AcceptHeader
func AcceptHeader(s string) Option {
	return func(c *Config) error {
		c.AcceptHeader = s
		return nil
	}
}

func validateBaseURLOption(url string) error {
	// Validate that there is no trailing slashes before setting the custom baseURL
	if url[len(url)-1:] == "/" {
		return fmt.Errorf("custom base URL cannot contain a trailing slash")
	}
	return nil
}
