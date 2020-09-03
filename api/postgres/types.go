package postgres

// MTLSConfigStatus represents the status of a MTLS configuration
type MTLSConfigStatus string

// MTLSConfigStatuses define all statuses pertaining to the lifecycle of a MTLS configuration.
var MTLSConfigStatuses = struct {
	PROVISIONING   MTLSConfigStatus
	DEPROVISIONING MTLSConfigStatus
	DEPROVISIONED  MTLSConfigStatus
	OPERATIONAL    MTLSConfigStatus
	UNKNOWN        MTLSConfigStatus
}{
	PROVISIONING:   "Provisioning",
	DEPROVISIONING: "Deprovisioning",
	DEPROVISIONED:  "Deprovisioned",
	OPERATIONAL:    "Operational",
	UNKNOWN:        "Unknown",
}

// ToString is a helper method to return the string of a MTLSConfigStatus.
func (s MTLSConfigStatus) ToString() string {
	return string(s)
}

// MTLSIPRuleStatus represent the status of a MTLS IP rule.
type MTLSIPRuleStatus string

// MTLSIPRuleStatuses define all statuses pertaining to the lifecycle of a MTLS IP config.
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
