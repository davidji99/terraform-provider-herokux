package api

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	"sync"
	"time"
)

const (
	// DefaultAPIBaseURL is the base URL.
	DefaultAPIBaseURL = "https://api.heroku.com"

	// MetricAPIBaseURL is the base Metric URL.
	MetricAPIBaseURL = "https://api.metrics.heroku.com"

	// DefaultUserAgent is the user agent used when making API calls.
	DefaultUserAgent = "herokux-go"

	// DefaultAcceptHeader is the default and required Accept header.
	DefaultAcceptHeader = "application/vnd.heroku+json; version=3"
)

// A Client manages communication with the Heroku API.
type Client struct {
	// clientMu protects the client during calls that modify the CheckRedirect func.
	clientMu sync.Mutex

	// HTTP client used to communicate with the API.
	http *simpleresty.Client

	// baseURL for API. No trailing slashes.
	baseURL string

	// Reuse a single struct instead of allocating one for each service on the heap.
	common service

	// User agent used when communicating with the Heroku API.
	userAgent string

	// Custom HTTPHeaders
	customHTTPHeaders map[string]string

	// API token
	apiToken string

	// Services used for talking to different parts of the Heroku API.
	Formations *FormationsService
}

// service represents the API service client.
type service struct {
	client *Client
}

// New constructs a new client to interact with the API.
func New(opts ...Option) (*Client, error) {
	// Construct new client.
	c := &Client{
		http:              simpleresty.New(),
		baseURL:           DefaultAPIBaseURL,
		userAgent:         DefaultUserAgent,
		customHTTPHeaders: map[string]string{},
		apiToken:          "",
	}

	// Define any user custom Client settings
	if optErr := c.parseOptions(opts...); optErr != nil {
		return nil, optErr
	}

	// Validate that apiToken is set on the Client
	if c.apiToken == "" {
		return nil, fmt.Errorf("no API token defined for this Client")
	}

	// Setup the client with default settings
	c.setupClient()

	// Inject services
	c.injectServices()

	return c, nil
}

// setupClient sets common headers and other configurations.
func (c *Client) setupClient() {
	// Set Base URL
	c.http.SetBaseURL(c.baseURL)

	/*
		We aren't setting an authentication header initially here as certain API resources require specific access_tokens.
		Per Heroku API documentation, each individual resource will set the access_token parameter when constructing
		the full API endpoint URL.
	*/
	c.http.SetHeader("Content-type", "application/json").
		SetHeader("Accept", DefaultAcceptHeader).
		SetHeader("User-Agent", c.userAgent).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.apiToken)).
		SetTimeout(1 * time.Minute).
		SetAllowGetMethodPayload(true)

	// Set additional headers
	if c.customHTTPHeaders != nil {
		c.http.SetHeaders(c.customHTTPHeaders)
	}
}

// injectServices adds the services to the client.
func (c *Client) injectServices() {
	c.common.client = c
	c.Formations = (*FormationsService)(&c.common)
}

// parseOptions parses the supplied options functions and returns a configured *Client instance.
func (c *Client) parseOptions(opts ...Option) error {
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
