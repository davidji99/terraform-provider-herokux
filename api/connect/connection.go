package connect

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	"time"
)

// AuthResponse represents the response detailing all connections belonging to the authenticated user.
type AuthResponse struct {
	User        *ConnectionUser           `json:"user,omitempty"`
	Connections []*AuthResponseConnection `json:"connections,omitempty"`
}

// AuthResponseConnection represent information returned about a Heroku Connect connection after an Connect Auth request.
type AuthResponseConnection struct {
	ID           *string `json:"id,omitempty"`
	ResourceName *string `json:"resource_name,omitempty"`
	DisplayName  *string `json:"display_name,omitempty"`
	AddonType    *string `json:"addon_type,omitempty"`
	AppName      *string `json:"app_name,omitempty"`
	AppID        *string `json:"app_id,omitempty"`
	Region       *string `json:"region,omitempty"`
	RegionURL    *string `json:"region_url,omitempty"`
	RegionFlag   *string `json:"region_flag,omitempty"`
	RegionLabel  *string `json:"region_label,omitempty"`
	CellLabel    *string `json:"cell_label,omitempty"`
	Cell         *string `json:"cell,omitempty"`
	DetailURL    *string `json:"detail_url,omitempty"`
}

// ConnectionUser represents the user for a connection.
type ConnectionUser struct {
	ID    *string `json:"id,omitempty"`
	Email *string `json:"email,omitempty"`
}

// Connection represent a Heroku Connect connection.
type Connection struct {
	ID                   *string                   `json:"id,omitempty"`
	Name                 *string                   `json:"name,omitempty"`
	AddonID              *string                   `json:"addon_id,omitempty"`
	AppName              *string                   `json:"app_name,omitempty"`
	TeamName             *string                   `json:"team_name,omitempty"`
	AppID                *string                   `json:"app_id,omitempty"`
	SchemaName           *string                   `json:"schema_name,omitempty"`
	DBKey                *string                   `json:"db_key,omitempty"`
	Database             *ConnectionDatabase       `json:"database,omitempty"`
	OrganizationID       *string                   `json:"organization_id,omitempty"`
	State                *string                   `json:"state,omitempty"`
	MappingsSummaryState *string                   `json:"mappings_summary_state,omitempty"`
	DetailURL            *string                   `json:"detail_url,omitempty"`
	Tags                 []string                  `json:"tags,omitempty"`
	SalesforceInfo       *ConnectionSalesforceInfo `json:"sf_info,omitempty"`
	CreatedAt            *time.Time                `json:"created_at,omitempty"`
	Plan                 *string                   `json:"plan,omitempty"`
	FreeEdition          *bool                     `json:"free_edition"`
	Dogwood              *bool                     `json:"dogwood"`
	LogplexLogEnabled    *bool                     `json:"logplex_log_enabled"`
	Deletable            *bool                     `json:"deletable"`
	ResourceName         *string                   `json:"resource_name,omitempty"`
	LargeQueryThreshold  *int                      `json:"large_query_threshold,omitempty"`
	BulkPageSize         *int                      `json:"bulk_page_size,omitempty"`
	Features             *ConnectionFeatures       `json:"features,omitempty"`
	InternalName         *string                   `json:"internal_name,omitempty"`
	NotificationUrl      *string                   `json:"notifications_url,omitempty"`
	SOAPBatchSize        *int                      `json:"soap_batch_size,omitempty"`
	SalesforceInstance   *string                   `json:"sf_instance,omitempty"`
	SalesforceRegion     *string                   `json:"sf_region,omitempty"`
	Mappings             []*Mapping                `json:"mappings,omitempty"`
	OData                *ConnectionOData          `json:"odata,omitempty"`
	// AuthUpdated TODO: Not sure what the data type is
	// UserInfo TODO: Not sure what the data type is
	// Metrics TODO: Not sure what the data type is
}

type ConnectionOData struct {
	ID   *int    `json:"id,omitempty"`
	Slug *string `json:"slug,omitempty"`
}

type ConnectionDatabase struct {
	Host *string `json:"host,omitempty"`
	Name *string `json:"database,omitempty"`
	Port *int    `json:"port,omitempty"`
}

type ConnectionSalesforceInfo struct {
	Username       *string `json:"username,omitempty"`
	OrganizationID *string `json:"organization_id,omitempty"`
	Environment    *string `json:"environment,omitempty"`
	Domain         *string `json:"domain,omitempty"`
	APIVersion     *string `json:"api_version,omitempty"`
}

type ConnectionFeatures struct {
	DbUpdatesForwardInTimeOnly *bool `json:"db_updates_forward_in_time_only"`
	DbaExceptionTranslation    *bool `json:"dba_exception_translation"`
	DisableBulkWrites          *bool `json:"disable_bulk_writes"`
	DriftDetectorBatchDisk     *bool `json:"drift_detector_batch_disk"`
	DriftDetectorStateMachine  *bool `json:"drift_detector_state_machine"`
	EarlySfApiAccess           *bool `json:"early_sf_api_access"`
	ForcePicklist255           *bool `json:"force_picklist_255"`
	MinimalIndexesByDefault    *bool `json:"minimal_indexes_by_default"`
	MTLSShieldConnection       *bool `json:"mtls_shield_connection"`
	PollDbNoMerge              *bool `json:"poll_db_no_merge"`
	PollExternalIDs            *bool `json:"poll_external_ids"`
	RestCountOnly              *bool `json:"rest_count_only"`
	SSHHostKeyVerification     *bool `json:"ssh_host_key_verification"`
	SyncRepairImmediately      *bool `json:"sync_repair_immediately"`
	TempTablesInTransactions   *bool `json:"temp_tables_in_transactions"`
	TestingFlagOnly            *bool `json:"testing_flag_only"`
	UserAssignmentRules        *bool `json:"use_assignment_rules"`
}

type ConnectionUpdateRequest struct {
	Name       string `json:"name,omitempty"`
	SchemaName string `json:"schema_name,omitempty"`
	DBKey      string `json:"db_key,omitempty"`
}

// ConnectionGetQueryParams are query parameters available when retrieving a connection.
//
// Reference: https://devcenter.heroku.com/articles/heroku-connect-api#step-7-monitor-the-connection-and-mapping-status
type ConnectionGetQueryParams struct {
	// Deep adds connection status and mapping status to the connection response body.
	Deep bool `url:"deep,omitempty"`
}

// GetConnection retrieves information about a connection.
func (c *Connect) GetConnection(connectionID string, params ConnectionGetQueryParams) (*Connection, *simpleresty.Response, error) {
	var result *Connection
	urlStr, urlStrErr := c.http.RequestURLWithQueryParams(fmt.Sprintf("/api/v3/connections/%s", connectionID), params)
	if urlStrErr != nil {
		return nil, nil, urlStrErr
	}

	// Execute the request
	response, updateErr := c.http.Get(urlStr, &result, nil)

	return result, response, updateErr
}

// ConfigureSettings updates a Heroku Connect connection.
//
// Reference: https://devcenter.heroku.com/articles/heroku-connect-api#step-4-configure-the-database-key-and-schema-for-the-connection
func (c *Connect) ConfigureSettings(connectionID string, opts *ConnectionUpdateRequest) (*Connection, *simpleresty.Response, error) {
	var result *Connection
	urlStr := c.http.RequestURL("/api/v3/connections/%s", connectionID)

	// Execute the request
	response, updateErr := c.http.Patch(urlStr, &result, opts)

	return result, response, updateErr
}
