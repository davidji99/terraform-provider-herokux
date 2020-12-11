package connect

import "github.com/davidji99/simpleresty"

// ConnectionCredential represents credential for a Heroku Connect connection.
type ConnectionCredential struct {
	ID                    *string                         `json:"id,omitempty"`
	Slug                  *string                         `json:"slug,omitempty"`
	Connection            *AuthResponseConnection         `json:"connection,omitempty"`
	Enabled               *bool                           `json:"enabled"`
	AccessURL             *string                         `json:"access_url,omitempty"`
	Resources             []*ConnectionCredentialResource `json:"resources,omitempty"`
	APIHourlyRateLimit    *int                            `json:"api_hourly_rate_limit,omitempty"`
	CurrentHourlyAPIUsage *int                            `json:"current_hourly_api_usage,omitempty"`
	Credentials           *ConnectionCredentialSecrets    `json:"credentials,omitempty"`

	// DatabaseStatementTimeOut
}

type ConnectionCredentialResource struct {
	ID         *int    `json:"id,omitempty"`
	Services   *int    `json:"services,omitempty"`
	Exported   *bool   `json:"exported,omitempty"`
	SchemaName *string `json:"schema_name,omitempty"`
	TableName  *string `json:"table_name,omitempty"`
	IsValid    *bool   `json:"is_valid,omitempty"`
	Type       *string `json:"type,omitempty"`
}

type ConnectionCredentialSecrets struct {
	User     *string `json:"user,omitempty"`
	Password *string `json:"password,omitempty"`
}

// CreateCredential creates a user/password for accessing your shared data sources on a connection.
func (c *Connect) CreateCredential(odataID string) (*ConnectionCredential, *simpleresty.Response, error) {
	var result *ConnectionCredential
	urlStr := c.http.RequestURL("/api/v3/odata/services/%s", odataID)

	opts := struct {
		Enabled bool `json:"enabled"`
	}{
		Enabled: true,
	}

	// Execute the request
	response, updateErr := c.http.Patch(urlStr, &result, opts)

	return result, response, updateErr
}

// RevokeCredential revokes (invalidates) a user/password for accessing your shared data sources on a connection.
func (c *Connect) RevokeCredential(odataID string) (*ConnectionCredential, *simpleresty.Response, error) {
	var result *ConnectionCredential
	urlStr := c.http.RequestURL("/api/v3/odata/services/%s", odataID)

	opts := struct {
		Enabled bool `json:"enabled"`
	}{
		Enabled: false,
	}

	// Execute the request
	response, updateErr := c.http.Patch(urlStr, &result, opts)

	return result, response, updateErr
}
