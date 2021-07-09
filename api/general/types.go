package general

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
