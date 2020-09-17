package postgres

import (
	"github.com/davidji99/simpleresty"
	"github.com/davidji99/terraform-provider-herokux/api/data"
)

// Privatelink represents a connection between a Heroku postgres and AWS resources.
//
// The fields `AllowedAccounts` and `WhitelistedAccounts` are set to the same values.
type Privatelink struct {
	App                 *PrivatelinkApp               `json:"app,omitempty"`
	Addon               *PrivatelinkAddon             `json:"addon,omitempty"`
	Status              *data.PrivatelinkStatus       `json:"status,omitempty"`
	ServiceName         *string                       `json:"service_name,omitempty"`
	AllowedAccounts     []*PrivatelinkAllowedAccounts `json:"allowed_accounts,omitempty"`
	WhitelistedAccounts []*PrivatelinkAllowedAccounts `json:"whitelisted_accounts,omitempty"`
	Connections         []*PrivatelinkConnections     `json:"connections,omitempty"`
}

// PrivatelinkAllowedAccounts represents AWS accounts granted access to a privatelink.
type PrivatelinkAllowedAccounts struct {
	AccountID *string                               `json:"account_id,omitempty"`
	ARN       *string                               `json:"arn,omitempty"`
	Status    *data.PrivatelinkAllowedAccountStatus `json:"status,omitempty"`
}

// PrivatelinkConnections represents connections between Heroku postgres and AWS resources.
type PrivatelinkConnections struct {
	EndpointID *string `json:"endpoint_id,omitempty"`
	Hostname   *string `json:"hostname,omitempty"`
	OwnerARN   *string `json:"owner_arn,omitempty"`
	Status     *string `json:"status,omitempty"`
}

// PrivatelinkApp represents the app hosting the addon that's has privatelink enabled.
type PrivatelinkApp struct {
	Name *string `json:"name,omitempty"`
}

// PrivatelinkAddon represents the addon that's being provisioned a privatelink.
type PrivatelinkAddon struct {
	Name *string `json:"name,omitempty"`
	UUID *string `json:"uuid,omitempty"`
}

// PrivatelinkRequest represents a request to create or update privatelink.
type PrivatelinkRequest struct {
	AllowedAccounts []string `json:"allowed_accounts,omitempty"`
}

// CreatePrivatelink creates a privatelink for a Heroku postgres, redis, or kafka addon.
func (p *Postgres) CreatePrivatelink(addonID string, opts *PrivatelinkRequest) (*Privatelink, *simpleresty.Response, error) {
	var result *Privatelink
	urlStr := p.http.RequestURL("/private-link/v0/databases/%s", addonID)

	// Execute the request
	response, createErr := p.http.Post(urlStr, &result, opts)

	return result, response, createErr
}

// GetPrivatelink gets information about a privatelink for a Heroku postgres, redis, or kafka addon.
func (p *Postgres) GetPrivatelink(addonID string) (*Privatelink, *simpleresty.Response, error) {
	var result *Privatelink
	urlStr := p.http.RequestURL("/private-link/v0/databases/%s", addonID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// DeletePrivatelink deletes a privatelink for a Heroku postgres, redis, or kafka addon.
//
// A successful DELETE requests results in the `status` set to Deprovisioned.
// However, the UI will show "deprovisioning" so user needs to use `GetPrivatelink` method in a loop
// until the GET request returns a `404`.
func (p *Postgres) DeletePrivatelink(addonID string) (*Privatelink, *simpleresty.Response, error) {
	var result *Privatelink
	urlStr := p.http.RequestURL("/private-link/v0/databases/%s", addonID)

	// Execute the request
	response, deleteErr := p.http.Delete(urlStr, &result, nil)

	return result, response, deleteErr
}

// RemovePrivatelinkAllowedAccounts removes one or more allowed accounts from a private link.
func (p *Postgres) RemovePrivatelinkAllowedAccounts(addonID string, opts *PrivatelinkRequest) (*Privatelink, *simpleresty.Response, error) {
	var result *Privatelink
	urlStr := p.http.RequestURL("/private-link/v0/databases/%s/allowed_accounts", addonID)

	// Execute the request
	response, patchErr := p.http.Patch(urlStr, &result, opts)

	return result, response, patchErr
}

// AddPrivatelinkAllowedAccounts adds one more allowed accounts to a privatelink.
//
// Warning: the UI may become a bit wonky until the allowed account becomes `Active`.
func (p *Postgres) AddPrivatelinkAllowedAccounts(addonID string, opts *PrivatelinkRequest) (*Privatelink, *simpleresty.Response, error) {
	var result *Privatelink
	urlStr := p.http.RequestURL("/private-link/v0/databases/%s/allowed_accounts", addonID)

	// Execute the request
	response, patchErr := p.http.Put(urlStr, &result, opts)

	return result, response, patchErr
}
