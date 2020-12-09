package postgres

import "github.com/davidji99/simpleresty"

// Credential represents a credential for a postgres database.
//
// The ID (UUID) cannot be used to retrieve a single credential. The name is required.
type Credential struct {
	ID          *string             `json:"uuid,omitempty"`
	Name        *string             `json:"name,omitempty"`
	State       CredentialState     `json:"state,omitempty"`
	Database    *string             `json:"database,omitempty"`
	Host        *string             `json:"host,omitempty"`
	Port        *int                `json:"port,omitempty"`
	Credentials []*CredentialSecret `json:"credentials,omitempty"`
}

// CredentialSecret represents the username & password for a Credential.
type CredentialSecret struct {
	User     *string `json:"user,omitempty"`
	Password *string `json:"password,omitempty"`
	State    *string `json:"state,omitempty"`
}

// ListCredentials retrieves all credentials for a database.
func (p *Postgres) ListCredentials(nameOrID string) ([]*Credential, *simpleresty.Response, error) {
	var result []*Credential
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/credentials", nameOrID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// GetCredential retrieves a single credential for a database.
func (p *Postgres) GetCredential(nameOrID, credentialName string) (*Credential, *simpleresty.Response, error) {
	var result *Credential
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/credentials/%s", nameOrID, credentialName)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// CreateCredential creates a postgres database credential.
//
// Returns a GenericResponse.
func (p *Postgres) CreateCredential(nameOrID, newCredName string) (*GenericResponse, *simpleresty.Response, error) {
	var result *GenericResponse
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/credentials", nameOrID)

	body := struct {
		Name string `json:"name"`
	}{
		Name: newCredName,
	}

	// Execute the request
	response, getErr := p.http.Post(urlStr, &result, &body)

	return result, response, getErr
}

// DeleteCredential revokes and deletes a credential. Please make sure to first check if the credential is attached
// to any existing addons.
//
// Returns a GenericResponse with a message of the following:
//  - The credential from_api has been destroyed within postgresql-fluffy-50793 and detached from all apps.
//
// Note: it takes a bit of time before the credential is fully deleted. The username/password are first to be deleted
// and then the credential itself is deleted.
func (p *Postgres) DeleteCredential(nameOrID, credentialName string) (*GenericResponse, *simpleresty.Response, error) {
	var result *GenericResponse
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/credentials/%s", nameOrID, credentialName)

	// Execute the request
	response, getErr := p.http.Delete(urlStr, &result, nil)

	return result, response, getErr
}
