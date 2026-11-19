// Package advanced accesses the Heroku Postgres Advanced data API.
package advanced

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	http *resty.Client
}

var ErrNotFound = errors.New("Postgres Advanced resource not found")

func New(c *config.Config) *Client {
	client := resty.New().SetBaseURL(c.PostgresAdvancedBaseURL).
		SetTimeout(2*time.Minute).
		SetHeader("Content-Type", c.ContentTypeHeader).
		SetHeader("Accept", c.AcceptHeader).
		SetHeader("User-Agent", c.UserAgent).
		SetHeader("Authorization", "Bearer "+c.APIToken).
		SetHeaders(c.CustomHTTPHeaders)
	return &Client{http: client}
}

func databasePath(id string) string {
	return "/data/postgres/v1/" + url.PathEscape(id)
}

func (c *Client) request(ctx context.Context, method, path string, body, result any) error {
	resp, err := c.http.R().SetContext(ctx).SetBody(body).Execute(method, path)
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusNotFound {
		return fmt.Errorf("%s %s: %w", method, path, ErrNotFound)
	}
	if !resp.IsSuccess() {
		var detail struct{ Message string }
		if json.Unmarshal(resp.Body(), &detail) == nil && detail.Message != "" {
			return fmt.Errorf("%s %s: %s: %s", method, path, resp.Status(), detail.Message)
		}
		return fmt.Errorf("%s %s: %s", method, path, resp.Status())
	}
	if result != nil {
		if err := json.Unmarshal(resp.Body(), result); err != nil {
			return fmt.Errorf("decoding %s %s: %w", method, path, err)
		}
	}
	return nil
}
