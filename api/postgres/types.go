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

// DatabaseInfoName represents a database info name.
type DatabaseInfoName string

// DatabaseInfoNames represents database infor names.
var DatabaseInfoNames = struct {
	PLAN       DatabaseInfoName
	STATUS     DatabaseInfoName
	HASTATUS   DatabaseInfoName
	DATASIZE   DatabaseInfoName
	PGVERSION  DatabaseInfoName
	FORKFOLLOW DatabaseInfoName
	REGION     DatabaseInfoName
	MUTUALTLS  DatabaseInfoName
	FOLLOWERS  DatabaseInfoName
}{
	PLAN:       "Plan",
	STATUS:     "Status",
	HASTATUS:   "HA Status",
	DATASIZE:   "Data Size",
	PGVERSION:  "PG Version",
	FORKFOLLOW: "Fork/Follow",
	REGION:     "Region",
	MUTUALTLS:  "Mutual TLS",
	FOLLOWERS:  "Followers",
}

// ToString is a helper method to return the string of a DatabaseInfoName.
func (s DatabaseInfoName) ToString() string {
	return string(s)
}

// DataConnectorStatus represents the status of a Data Connector.
type DataConnectorStatus string

// DataConnectorStatuses represent all statuses pertaining to the lifecycle of a data connector.
var DataConnectorStatuses = struct {
	CREATING      DataConnectorStatus
	AVAILABLE     DataConnectorStatus
	DEPROVISIONED DataConnectorStatus
	DELETED       DataConnectorStatus
	PAUSED        DataConnectorStatus
	UNKNOWN       DataConnectorStatus
}{
	CREATING:      "creating",
	AVAILABLE:     "available",
	DELETED:       "DELETED",
	DEPROVISIONED: "deprovisioned",
	PAUSED:        "paused",
	UNKNOWN:       "unknown",
}

// ToString is a helper method to return the string of a DataConnectorStatus.
func (s DataConnectorStatus) ToString() string {
	return string(s)
}

// CredentialState represents the status of a postgres credential.
type CredentialState string

// CredentialStates represent all statuses pertaining to the lifecycle of a postgres credential.
var CredentialStates = struct {
	PROVISIONING        CredentialState
	WAITFORPROVISIONING CredentialState
	ACTIVE              CredentialState
	REVOKING            CredentialState
	DELETED             CredentialState
	UNKNOWN             CredentialState
}{
	ACTIVE:              "active",
	WAITFORPROVISIONING: "wait_for_provisioning",
	PROVISIONING:        "provisioning",
	REVOKING:            "revoking",
	DELETED:             "DELETED",
	UNKNOWN:             "unknown",
}

// ToString is a helper method to return the string of a CredentialState.
func (s CredentialState) ToString() string {
	return string(s)
}

// LogStatementUpdateOption represents the option to update log statement setting.
type LogStatementUpdateOption string

// LogStatementUpdateOptions represents all options when updating log statements.
var LogStatementUpdateOptions = struct {
	// None: No statements are logged.
	None LogStatementUpdateOption

	// DDL: All data definition statements, such as CREATE, ALTER and DROP will be logged.
	DDL LogStatementUpdateOption

	// Mod: Includes all statements from ddl as well as data-modifying statements such as INSERT, UPDATE, DELETE, TRUNCATE, COPY.
	Mod LogStatementUpdateOption

	// All:  statements are logged.
	All LogStatementUpdateOption
}{
	None: "none",
	DDL:  "ddl",
	Mod:  "mod",
	All:  "all",
}
