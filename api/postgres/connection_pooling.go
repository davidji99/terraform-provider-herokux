package postgres

import (
	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/simpleresty"
)

// ConnectionPoolingRequest represents
type ConnectionPoolingRequest struct {
	// Name of config var
	Name string `json:"name"`

	// Credential
	Credential string `json:"credential"`

	// App name or UUID
	App string `json:"app"`
}

// CreateConnectionPooling activates connection pooling for a database.
func (p *Postgres) CreateConnectionPooling(nameOrID string, opts *ConnectionPoolingRequest) (*heroku.AddOnAttachment, *simpleresty.Response, error) {
	var result heroku.AddOnAttachment

	urlStr := p.http.RequestURL("/client/v11/databases/%s/connection-pooling", nameOrID)

	// Execute the request
	response, createErr := p.http.Post(urlStr, &result, opts)

	return &result, response, createErr
}
