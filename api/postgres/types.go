package postgres

// MTLSConfigStatus represents the status of a MTLS configuration
type MTLSConfigStatus string

// MTLSConfigStatuses represent all statuses pertaining to the lifecycle of a MTLS configuration.
var MTLSConfigStatuses = struct {
	PROVISIONING   MTLSConfigStatus
	DEPROVISIONING MTLSConfigStatus
	DEPROVISIONED  MTLSConfigStatus
	OPERATIONAL    MTLSConfigStatus
	UNKNOWN        MTLSConfigStatus
	SERVERERROR    MTLSConfigStatus // Represents when GETing the status of the MTLS provisioning sometimes returns a 500
}{
	PROVISIONING:   "Provisioning",
	DEPROVISIONING: "Deprovisioning",
	DEPROVISIONED:  "Deprovisioned",
	OPERATIONAL:    "Operational",
	UNKNOWN:        "Unknown",
	SERVERERROR:    "ServerError",
}

// ToString is a helper method to return the string of a MTLSConfigStatus.
func (s MTLSConfigStatus) ToString() string {
	return string(s)
}

// MTLSIPRuleStatus represent the status of a MTLS IP rule.
type MTLSIPRuleStatus string

// MTLSIPRuleStatuses represent all statuses pertaining to the lifecycle of a MTLS IP config.
var MTLSIPRuleStatuses = struct {
	AUTHORIZING MTLSIPRuleStatus
	AUTHORIZED  MTLSIPRuleStatus
	UNKNOWN     MTLSIPRuleStatus
}{
	AUTHORIZED:  "Authorized",
	AUTHORIZING: "Authorizing",
	UNKNOWN:     "Unknown",
}

// ToString is a helper method to return the string of a MTLSIPRuleStatus.
func (s MTLSIPRuleStatus) ToString() string {
	return string(s)
}

// MTLSCertStatus represent the status of a MTLS certificate.
type MTLSCertStatus string

// MTLSCertStatuses represent all statuses pertaining to the lifecycle of a MTLS certificate.
var MTLSCertStatuses = struct {
	READY     MTLSCertStatus
	PENDING   MTLSCertStatus
	DISABLING MTLSCertStatus
	DISABLED  MTLSCertStatus
	UNKNOWN   MTLSCertStatus
}{
	READY:     "ready",
	PENDING:   "pending",
	DISABLING: "disabling",
	DISABLED:  "disabled",
	UNKNOWN:   "unknown",
}

// ToString is a helper method to return the string of a MTLSCertStatus.
func (s MTLSCertStatus) ToString() string {
	return string(s)
}
