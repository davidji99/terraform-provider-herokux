package general

import "time"

// MtlsIPRule represents a MTLS IP rule.
type MtlsIPRule struct {
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
