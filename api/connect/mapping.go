package connect

import (
	"encoding/json"
	"github.com/davidji99/simpleresty"
	"time"
)

// MappingExportOutput represents the raw output from exporting mappings on a connection.
type MappingExportOutput map[string]interface{}

// ToString converts the mapping export output to a string value.
func (m *MappingExportOutput) ToString() (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// ToByteArray converts the mapping export output to a []byte value.
func (m *MappingExportOutput) ToByteArray() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Mapping represents a connection mapping.
type Mapping struct {
	ID                  *string                 `json:"id,omitempty"`
	ObjectName          *string                 `json:"object_name,omitempty"`
	State               *string                 `json:"state,omitempty"`
	Config              *MappingConfig          `json:"config,omitempty"`
	DetailURL           *string                 `json:"detail_url,omitempty"`
	ErrorsURL           *string                 `json:"errors_url,omitempty"`
	SchemaURL           *string                 `json:"schema_url,omitempty"`
	UpdatedAt           *time.Time              `json:"updated_at,omitempty"`
	CreatedAt           *time.Time              `json:"created_at,omitempty"`
	Connection          *AuthResponseConnection `json:"connection,omitempty"`
	SFNotifyEnabled     *bool                   `json:"sf_notify_enabled"`
	SFPollingSeconds    *int                    `json:"sf_polling_seconds,omitempty"`
	Access              *string                 `json:"access,omitempty"`
	Counts              *MappingCounts          `json:"counts,omitempty"`
	ActivelyWriting     *bool                   `json:"actively_writing,omitempty"`
	StateDescription    *string                 `json:"state_description,omitempty"`
	HasHCColumns        *bool                   `json:"has_hc_columns,omitempty"`
	DBCountDate         *time.Time              `json:"db_count_date,omitempty"`
	SalesforceCountDate *time.Time              `json:"salesforce_count_date,omitempty"`
	IsPolling           *bool                   `json:"is_polling,omitempty"`
	// UpsertField
	// Times
	// SyncFlags
}

type MappingConfig struct {
	Access             *string `json:"access,omitempty"`
	SFNotifyEnabled    *bool   `json:"sf_notify_enabled"`
	SFPollingSeconds   *int    `json:"sf_polling_seconds,omitempty"`
	SFMaxDailyAPICalls *int    `json:"sf_max_daily_api_calls,omitempty"`
	// Fields
	// Indexes
	Revision  *int       `json:"revision,omitempty"`
	AppliedAt *time.Time `json:"applied_at,omitempty"`
}

type MappingCounts struct {
	Count   *int `json:"count,omitempty"`
	SF      *int `json:"sf,omitempty"`
	Pending *int `json:"pending,omitempty"`
	Errors  *int `json:"errors,omitempty"`
}

// MappingsExport represents the data structure returned when exporting an existing connection's mappings.
type MappingsExport struct {
	Mappings   []*Mapping  `json:"mappings"`
	Connection *Connection `json:"connection"`
	Version    *int        `json:"version"`
}

// GetMapping retrieves information about a connection mapping.
func (c *Connect) GetMapping(mappingID string) (*Mapping, *simpleresty.Response, error) {
	var result *Mapping
	urlStr := c.http.RequestURL("/api/v3/mappings/%s", mappingID)

	// Execute the request
	response, updateErr := c.http.Get(urlStr, &result, nil)

	return result, response, updateErr
}

// ImportMappings takes a JSON file path and creates mappings from it.
func (c *Connect) ImportMappings(connectID string, mappings []byte) (*simpleresty.Response, error) {
	urlStr := c.http.RequestURL("/api/v3/connections/%s/actions/import", connectID)

	// Execute the request
	response, updateErr := c.http.Post(urlStr, nil, mappings)

	return response, updateErr
}

// ExportMappings exports mappings for an existing connection.
func (c *Connect) ExportMappings(connectID string) (*MappingExportOutput, *simpleresty.Response, error) {
	var result *MappingExportOutput
	urlStr := c.http.RequestURL("/api/v3/connections/%s/actions/export", connectID)

	// Execute the request.
	// Result is a map[string]interface{}.
	response, updateErr := c.http.Get(urlStr, &result, nil)

	return result, response, updateErr
}

// DeleteMapping deletes a connection mapping.
func (c *Connect) DeleteMapping(mappingID string) (*simpleresty.Response, error) {
	var result *Mapping
	urlStr := c.http.RequestURL("/api/v3/mappings/%s", mappingID)

	// Execute the request
	response, updateErr := c.http.Delete(urlStr, &result, nil)

	return response, updateErr
}
