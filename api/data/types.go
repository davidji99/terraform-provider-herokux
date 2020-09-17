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

// PrivatelinkAllowedAccountStatus represents the status of a privatelink allowed account.
type PrivatelinkAllowedAccountStatus string

// PrivatelinkAllowedAccountStatuses represent all statuses pertaining to the lifecycle of a privatelink allowed account.
var PrivatelinkAllowedAccountStatuses = struct {
	PROVISIONING PrivatelinkAllowedAccountStatus
	ACTIVE       PrivatelinkAllowedAccountStatus
}{
	PROVISIONING: "Provisioning",
	ACTIVE:       "Active",
}

// ToString is a helper method to return the string of a PrivatelinkAllowedAccountStatus.
func (s PrivatelinkAllowedAccountStatus) ToString() string {
	return string(s)
}
