package postgres

import (
	"github.com/davidji99/simpleresty"
	"time"
)

// MTLSIPRule represents a MTLS IP rule.
type MTLSIPRule struct {
	ID          *string           `json:"id,omitempty"`
	CIDR        *string           `json:"cidr,omitempty"`
	Description *string           `json:"description,omitempty"`
	Status      *MTLSIPRuleStatus `json:"status,omitempty"`
	CreatedAt   *time.Time        `json:"created_at,omitempty"`
	UpdatedAt   *time.Time        `json:"updated_at,omitempty"`
}

// MTLSIPRuleRequest represents a request to create an IP rule.
type MTLSIPRuleRequest struct {
	CIDR        string `json:"cidr,omitempty"`
	Description string `json:"description,omitempty"`
}

func (p *Postgres) ListMTLSIPRules(dbNameOrID string) ([]*MTLSIPRule, *simpleresty.Response, error) {
	var result []*MTLSIPRule
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/ip-rules", dbNameOrID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

func (p *Postgres) GetMTLSIPRule(dbNameOrID, ipRuleID string) (*MTLSIPRule, *simpleresty.Response, error) {
	var result *MTLSIPRule
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/ip-rules/%s", dbNameOrID, ipRuleID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

func (p *Postgres) CreateMTLSIPRule(dbNameOrID string, opts *MTLSIPRuleRequest) (*MTLSIPRule, *simpleresty.Response, error) {
	var result *MTLSIPRule
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/ip-rules", dbNameOrID)

	// Execute the request
	response, createErr := p.http.Post(urlStr, &result, opts)

	return result, response, createErr
}

func (p *Postgres) DeleteMTLSIPRule(dbNameOrID, ipRuleID string) (*simpleresty.Response, error) {
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/tls-endpoint/ip-rules/%s", dbNameOrID, ipRuleID)

	// Execute the request
	response, deleteErr := p.http.Delete(urlStr, nil, nil)

	return response, deleteErr
}
