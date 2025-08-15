package postgres

import (
	"github.com/davidji99/simpleresty"
)

// MTLSCert represents a MTLS certificate.
type MTLSCert struct {
	ID                   *string         `json:"id,omitempty"`
	Name                 *string         `json:"name,omitempty"`
	CreatedAt            *string         `json:"created_at,omitempty"`
	UpdatedAt            *string         `json:"updated_at,omitempty"`
	ExpiresAt            *string         `json:"expires_at,omitempty"`
	Status               *MTLSCertStatus `json:"status,omitempty"`
	PrivateKey           *string         `json:"private_key,omitempty"`
	CertificateWithChain *string         `json:"certificate_with_chain,omitempty"`
}

// ListMTLSCerts lists all certificates.
//
// The certificates returned by this endpoint do not have their private keys and certificate chains in the response.
// To retrieve the key and chain, you must use the `GetMTLSCert` method.
// Furthermore, this endpoint also returns certificates that were disabled.
func (p *Postgres) ListMTLSCerts(dbNameOrID string) ([]*MTLSCert, *simpleresty.Response, error) {
	var result []*MTLSCert
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/certificates", dbNameOrID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// GetMTLSCert retrieves a single MTLS certificate.
//
// This endpoint returns a 404 if you retrieve a certificate that has been disabled/deleted.
func (p *Postgres) GetMTLSCert(dbNameOrID, certID string) (*MTLSCert, *simpleresty.Response, error) {
	var result *MTLSCert
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/certificates/%s", dbNameOrID, certID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// CreateMTLSCert creates a MTLS certificate.
//
// Upon creation, the new certificate has a status of 'pending'. A status of 'ready' signifies
// the certificate is ready for use.
func (p *Postgres) CreateMTLSCert(dbNameOrID string) (*MTLSCert, *simpleresty.Response, error) {
	var result *MTLSCert
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/certificates", dbNameOrID)

	// Execute the request
	response, createErr := p.http.Post(urlStr, &result, nil)

	return result, response, createErr

}

// DeleteMTLSCert deletes a MTLS certificate.
//
// Upon deletion, the target certificate has a status of 'disabling'.
func (p *Postgres) DeleteMTLSCert(dbNameOrID, certID string) (*MTLSCert, *simpleresty.Response, error) {
	var result *MTLSCert
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/certificates/%s", dbNameOrID, certID)

	// Execute the request
	response, deleteErr := p.http.Delete(urlStr, &result, nil)

	return result, response, deleteErr
}
