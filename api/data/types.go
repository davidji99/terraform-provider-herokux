package data

// PrivatelinkStatus represents the status of a privatelink.
type PrivatelinkStatus string

// PrivatelinkStatuses represent all statuses pertaining to the lifecycle of a private link.
var PrivatelinkStatuses = struct {
	PROVISIONING   PrivatelinkStatus
	OPERATIONAL    PrivatelinkStatus
	DEPROVISIONING PrivatelinkStatus
	DEPROVISIONED  PrivatelinkStatus
}{
	PROVISIONING:   "Provisioning",
	OPERATIONAL:    "Operational",
	DEPROVISIONING: "Deprovisioning",
	DEPROVISIONED:  "Deprovisioned", // Status is set to this even though the UI says deprovisioning. Must wait for 404.
}

// ToString is a helper method to return the string of a PrivatelinkStatus.
func (s PrivatelinkStatus) ToString() string {
	return string(s)
}
