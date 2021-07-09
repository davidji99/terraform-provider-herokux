package postgres

import (
	"github.com/davidji99/simpleresty"
	"github.com/davidji99/terraform-provider-herokux/api/general"
)

// ListMTLSIPRules returns all IP rules.
func (p *Postgres) ListMTLSIPRules(dbNameOrID string) ([]*general.MtlsIPRule, *simpleresty.Response, error) {
	var result []*general.MtlsIPRule
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/ip-rules", dbNameOrID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// GetMTLSIPRule returns a single IP rule.
func (p *Postgres) GetMTLSIPRule(dbNameOrID, ipRuleID string) (*general.MtlsIPRule, *simpleresty.Response, error) {
	var result *general.MtlsIPRule
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/ip-rules/%s", dbNameOrID, ipRuleID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// CreateMTLSIPRule creates an IP rule.
func (p *Postgres) CreateMTLSIPRule(dbNameOrID string, opts *general.MTLSIPRuleRequest) (*general.MtlsIPRule, *simpleresty.Response, error) {
	var result *general.MtlsIPRule
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/ip-rules", dbNameOrID)

	// Execute the request
	response, createErr := p.http.Post(urlStr, &result, opts)

	return result, response, createErr
}

// DeleteMTLSIPRule deletes an IP rule.
func (p *Postgres) DeleteMTLSIPRule(dbNameOrID, ipRuleID string) (*simpleresty.Response, error) {
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/ip-rules/%s", dbNameOrID, ipRuleID)

	// Execute the request
	response, deleteErr := p.http.Delete(urlStr, nil, nil)

	return response, deleteErr
}
