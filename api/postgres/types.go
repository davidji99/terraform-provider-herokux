package postgres

// MTLSConfigStatus represents the status of a MTLS configuration
type MTLSConfigStatus string

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

func (s MTLSConfigStatus) ToString() string {
	return string(s)
}
