package postgres

import (
	"github.com/davidji99/simpleresty"
)

// MTLSEndpoint represents the MTLS configuration for a given Heroku Postgres addon
type MTLSEndpoint struct {
	App                       *string           `json:"app,omitempty"`
	Addon                     *string           `json:"addon,omitempty"`
	Status                    *MTLSConfigStatus `json:"status,omitempty"`
	EnabledBy                 *string           `json:"enabled_by,omitempty"`
	CertificateAuthorityChain *string           `json:"certificate_authority_chain,omitempty"`
	//ActiveIPRules []string `json:"active_ip_rules,omitempty"`
}

// ProvisionMTLS enables MTLS for a database.
//
// If the request is successful, the response status code is 201 and the "status" is set to "Provisioning".
// Once the configuration is ready, "status" changes to "Operational".
func (p *Postgres) ProvisionMTLS(nameOrID string) (*MTLSEndpoint, *simpleresty.Response, error) {
	var result *MTLSEndpoint
	urlStr := p.http.RequestURL("/databases/%s/tls-endpoint", nameOrID)

	// Execute the request
	response, createErr := p.http.Post(urlStr, &result, nil)

	return result, response, createErr
}

// IsMTLSReady determines if the MTLS configuration is provisioned and operational.
//
// Return true if ready; false otherwise.
func (p *Postgres) IsMTLSReady(nameOrID string) (bool, MTLSConfigStatus, error) {
	mtlsConfig, _, getErr := p.GetMTLS(nameOrID)
	if getErr != nil {
		return false, MTLSConfigStatuses.UNKNOWN, getErr
	}

	if mtlsConfig.GetStatus() == &MTLSConfigStatuses.OPERATIONAL {
		return true, MTLSConfigStatuses.OPERATIONAL, nil
	}

	return false, MTLSConfigStatuses.PROVISIONING, nil
}

// DeprovisionMTLS destroys a MTLS configuration on your database.
//
// Returns 202 if request is successful with a 'status' of 'Deprovisioning'.
func (p *Postgres) DeprovisionMTLS(nameOrID string) (*MTLSEndpoint, *simpleresty.Response, error) {
	var result *MTLSEndpoint
	urlStr := p.http.RequestURL("/databases/%s/tls-endpoint", nameOrID)

	// Execute the request
	response, deleteErr := p.http.Delete(urlStr, &result, nil)

	return result, response, deleteErr
}

// GetMTLS retrieves the MTLS configuration for a database.
func (p *Postgres) GetMTLS(nameOrID string) (*MTLSEndpoint, *simpleresty.Response, error) {
	var result *MTLSEndpoint
	urlStr := p.http.RequestURL("/databases/%s/tls-endpoint", nameOrID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}
