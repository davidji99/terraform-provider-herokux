package advanced

import (
	"context"
	"net/http"
	"net/url"
)

type PublicationTarget struct {
	Type    string   `json:"type"`
	Tables  []string `json:"tables,omitempty"`
	Schemas []string `json:"schemas,omitempty"`
}

type Publication struct {
	Name   string            `json:"name"`
	Target PublicationTarget `json:"target"`
}

func publicationsPath(database string) string {
	return databasePath(database) + "/logical-replication/publications"
}

func (c *Client) CreatePublication(ctx context.Context, database string, pub Publication) error {
	return c.request(ctx, http.MethodPost, publicationsPath(database), pub, nil)
}

func (c *Client) GetPublication(ctx context.Context, database, name string) (Publication, error) {
	var pub Publication
	err := c.request(ctx, http.MethodGet, publicationsPath(database)+"/"+url.PathEscape(name), nil, &pub)
	return pub, err
}

func (c *Client) UpdatePublication(ctx context.Context, database, name string, target PublicationTarget) error {
	return c.request(ctx, http.MethodPut, publicationsPath(database)+"/"+url.PathEscape(name), map[string]PublicationTarget{"target": target}, nil)
}

func (c *Client) DeletePublication(ctx context.Context, database, name string) error {
	return c.request(ctx, http.MethodDelete, publicationsPath(database)+"/"+url.PathEscape(name), nil, nil)
}
