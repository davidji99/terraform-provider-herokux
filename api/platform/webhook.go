package platform

import (
	"github.com/davidji99/simpleresty"
	"time"
)

const (
	HerokuWebhookSecret = "Heroku-Webhook-Secret"
)

// AppWebhook enable you to receive notifications whenever particular changes are made to your Heroku app.
type AppWebhook struct {
	CreatedAt  *time.Time         `json:"created_at,omitempty"`
	ID         *string            `json:"id,omitempty"`
	EventTypes []WebhookEventType `json:"include,omitempty"`
	Level      *WebhookLevel      `json:"level,omitempty"`
	UpdatedAt  *time.Time         `json:"updated_at,omitempty"`
	URL        *string            `json:"url,omitempty"`
	Name       *string            `json:"name,omitempty"`
	App        *App               `json:"app,omitempty"`
}

// AppWebhookRequest represents a request to create or update a webhook.
type AppWebhookRequest struct {
	// Must be either notify or sync. If notify, Heroku does not retry unsuccessful webhook notifications.
	// If sync, Heroku retries failed requests until they succeed or until a determined limit is reached.
	Level WebhookLevel `json:"level,omitempty"`

	// The URL of your server endpoint that will receive all webhook notifications.
	URL string `json:"url,omitempty"`

	// List of the entities you want to subscribe to notifications for. Possible values are listed in the table
	// in the previous step. Any updates to this field will overwrite not append new values.
	Include []WebhookEventType `json:"include,omitempty"`

	// A value that Heroku will use to sign all webhook notification requests (the signature is included in the request’s
	// Heroku-Webhook-Hmac-SHA256 header). If you omit this value, a generated secret is returned by the CLI.
	// This value can never be retrieved again, so you must persist it immediately.
	Secret string `json:"secret,omitempty"`

	// Webhook name
	Name string `json:"name,omitempty"`
}

// ListAppWebhooks lists all webhooks for an app.
func (p *Platform) ListAppWebhooks(appID string) ([]*AppWebhook, *simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	var result []*AppWebhook
	urlStr := p.http.RequestURL("/apps/%s/webhooks", appID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", WebhooksAcceptHeader)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// GetAppWebhook retrieves a single webhook.
func (p *Platform) GetAppWebhook(appID, webhookID string) (*AppWebhook, *simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	var result *AppWebhook
	urlStr := p.http.RequestURL("/apps/%s/webhooks/%s", appID, webhookID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", WebhooksAcceptHeader)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// CreateAppWebhook creates an app webhook.
func (p *Platform) CreateAppWebhook(appID string, opts *AppWebhookRequest) (*AppWebhook, *simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	var result *AppWebhook
	urlStr := p.http.RequestURL("/apps/%s/webhooks", appID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", WebhooksAcceptHeader)

	// Execute the request
	response, getErr := p.http.Post(urlStr, &result, opts)

	return result, response, getErr
}

// UpdateAppWebhook modifies an existing app webhook.
func (p *Platform) UpdateAppWebhook(appID, webhookID string, opts *AppWebhookRequest) (*AppWebhook, *simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	var result *AppWebhook
	urlStr := p.http.RequestURL("/apps/%s/webhooks/%s", appID, webhookID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", WebhooksAcceptHeader)

	// Execute the request
	response, getErr := p.http.Patch(urlStr, &result, opts)

	return result, response, getErr
}

// DeleteAppWebhook deletes an app webhook.
func (p *Platform) DeleteAppWebhook(appID, webhookID string) (*simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	urlStr := p.http.RequestURL("/apps/%s/webhooks/%s", appID, webhookID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", WebhooksAcceptHeader)

	// Execute the request
	response, getErr := p.http.Delete(urlStr, nil, nil)

	return response, getErr
}
