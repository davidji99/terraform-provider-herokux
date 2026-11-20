// Package applink implements the Salesforce connection operations used by the Heroku AppLink CLI.
package applink

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/go-uuid"
)

var ErrNotFound = errors.New("AppLink connection not found")

type Client struct {
	config *config.Config
	http   *http.Client
}

func New(cfg *config.Config, transport http.RoundTripper) *Client {
	return &Client{config: cfg, http: &http.Client{
		Transport: transport, Timeout: 2 * time.Minute,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// request never puts response bodies, credentials, or request headers in diagnostics.
func (c *Client) request(ctx context.Context, method, endpoint string, headers http.Header, body, result any, connectionLookup bool) (http.Header, error) {
	var payload bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			return nil, fmt.Errorf("encoding AppLink request")
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, &payload)
	if err != nil {
		return nil, fmt.Errorf("invalid AppLink request URL")
	}
	req.Header = headers.Clone()
	req.Header.Set("User-Agent", c.config.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("%s %s: transport error contacting %s", method, req.URL.EscapedPath(), req.URL.Hostname())
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if connectionLookup && resp.StatusCode == http.StatusNotFound {
			var detail struct {
				ID string `json:"id"`
			}
			if json.NewDecoder(resp.Body).Decode(&detail) == nil && detail.ID == "record_not_found" {
				return nil, ErrNotFound
			}
		}
		return nil, fmt.Errorf("%s %s: HTTP %d", method, req.URL.EscapedPath(), resp.StatusCode)
	}
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return nil, fmt.Errorf("invalid response from %s %s", method, req.URL.EscapedPath())
		}
	}
	return resp.Header, nil
}

type Session struct {
	client  *Client
	origin  string
	addonID string
	headers http.Header
}

func trustedOrigin(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.User != nil {
		return ""
	}
	host := strings.ToLower(strings.TrimSuffix(u.Host, ":443"))
	switch host {
	case "heroku-applink.heroku.com", "applink-api.heroku.com":
		return "https://" + host
	default:
		return ""
	}
}

// Session resolves the attachment on the chosen app. Its headers belong only to the AppLink origin.
func (c *Client) Session(ctx context.Context, appID, addonID string) (*Session, error) {
	appURL := c.config.PlatformBaseURL + "/apps/" + url.PathEscape(appID)
	addonURL := appURL + "/addons/" + url.PathEscape(addonID)
	headers := make(http.Header)
	for k, v := range c.config.CustomHTTPHeaders {
		headers.Set(k, v)
	}
	headers.Set("Authorization", "Bearer "+c.config.APIToken)
	headers.Set("Accept", c.config.AcceptHeader)
	var addon heroku.AddOn
	if _, err := c.request(ctx, http.MethodGet, addonURL, headers, nil, &addon, false); err != nil {
		return nil, err
	}
	if addon.ID != addonID || addon.AddonService.Name != "heroku-applink" {
		return nil, fmt.Errorf("add-on %s is not a Heroku AppLink add-on on app %s", addonID, appID)
	}
	alias, err := c.attachmentAlias(ctx, appURL, appID, addonID, headers)
	if err != nil {
		return nil, err
	}
	var vars map[string]string
	if _, err = c.request(ctx, http.MethodGet, appURL+"/config-vars", headers, nil, &vars, false); err != nil {
		return nil, err
	}
	origin := trustedOrigin(vars[alias+"_API_URL"])
	if origin == "" {
		return nil, fmt.Errorf("attachment %s must identify a trusted AppLink origin", alias)
	}
	token := vars[alias+"_TOKEN"]
	if token == "" {
		return nil, fmt.Errorf("missing AppLink token for attachment %s", alias)
	}
	headers.Set("Accept", "application/vnd.heroku+json; version=3.sdk")
	var sso map[string]json.RawMessage
	if _, err = c.request(ctx, http.MethodGet, addonURL+"/sso", headers, nil, &sso, false); err != nil {
		return nil, err
	}
	if len(sso) == 0 {
		return nil, errors.New("Heroku returned an empty add-on SSO payload")
	}
	data, err := json.Marshal(sso)
	if err != nil {
		return nil, errors.New("could not encode add-on SSO payload")
	}
	serviceHeaders := make(http.Header)
	serviceHeaders.Set("Accept", "application/json")
	serviceHeaders.Set("Authorization", "Bearer "+token)
	serviceHeaders.Set("X-App-UUID", appID)
	serviceHeaders.Set("X-Addon-SSO", base64.StdEncoding.EncodeToString(data))
	return &Session{client: c, origin: origin, addonID: addonID, headers: serviceHeaders}, nil
}

func (c *Client) attachmentAlias(ctx context.Context, appURL, appID, addonID string, headers http.Header) (string, error) {
	page := headers.Clone()
	for {
		var attachments []heroku.AddOnAttachment
		h, err := c.request(ctx, http.MethodGet, appURL+"/addon-attachments", page, nil, &attachments, false)
		if err != nil {
			return "", err
		}
		for _, a := range attachments {
			if a.App.ID == appID && a.Addon.ID == addonID && a.Name != "" {
				return a.Name, nil
			}
		}
		next := h.Get("Next-Range")
		if next == "" {
			return "", fmt.Errorf("no attachment for AppLink add-on %s on app %s", addonID, appID)
		}
		if next == page.Get("Range") {
			return "", errors.New("Heroku repeated the attachment pagination range")
		}
		page.Set("Range", next)
	}
}

type Connection struct {
	ID      string `json:"id"`
	AddonID string `json:"addon_id"`
	Status  string `json:"status"`
	Org     struct {
		ID          string `json:"id"`
		Name        string `json:"connection_name"`
		Type        string `json:"type"`
		InstanceURL string `json:"instance_url"`
	} `json:"org"`
}

func (s *Session) path() string {
	return s.origin + "/addons/" + url.PathEscape(s.addonID) + "/connections/"
}

func (s *Session) Create(ctx context.Context, name, clientID, username, loginURL, key string) (Connection, error) {
	alias, err := uuid.GenerateUUID()
	if err != nil {
		return Connection{}, err
	}
	body := map[string]string{"alias": alias, "connection_name": name, "client_id": clientID, "username": username, "login_url": loginURL, "jwt_private_key": key}
	var conn Connection
	_, err = s.client.request(ctx, http.MethodPost, s.path()+"salesforce/jwt", s.headers, body, &conn, false)
	return conn, err
}

func (s *Session) Get(ctx context.Context, name string) (Connection, error) {
	var conn Connection
	_, err := s.client.request(ctx, http.MethodGet, s.path()+url.PathEscape(name), s.headers, nil, &conn, true)
	return conn, err
}

func (s *Session) Disconnect(ctx context.Context, name string) error {
	_, err := s.client.request(ctx, http.MethodDelete, s.path()+url.PathEscape(name), s.headers, nil, nil, true)
	return err
}
