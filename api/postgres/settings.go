package postgres

import "github.com/davidji99/simpleresty"

// Settings represents the available settings for a Heroku postgres database.
type Settings struct {
	LogLockWaits            *LogLockWaits            `json:"log_lock_waits,omitempty"`
	LogConnections          *LogConnections          `json:"log_connections,omitempty"`
	LogMinDurationStatement *LogMinDurationStatement `json:"log_min_duration_statement,omitempty"`
	LogStatement            *LogStatement            `json:"log_statement,omitempty"`
}

// LogLockWaits enables whether a log message is produced when a deadlock occurs.
type LogLockWaits struct {
	Value       *bool   `json:"value,omitempty"`
	Description *string `json:"desc,omitempty"`
	Default     *bool   `json:"default"`
}

// LogConnections enables logging of all attempted connection.
type LogConnections struct {
	Value       *bool   `json:"value,omitempty"`
	Description *string `json:"desc,omitempty"`
	Default     *bool   `json:"default"`
}

// LogMinDurationStatement causes the duration of each completed statement to be logged
// if the statement ran for at least the specified number of milliseconds.
type LogMinDurationStatement struct {
	Value       *int    `json:"value,omitempty"`
	Description *string `json:"desc,omitempty"`
	Default     *int    `json:"default"`
}

// LogStatement controls which SQL statements are logged.
type LogStatement struct {
	Value       *string             `json:"value,omitempty"`
	Description *string             `json:"desc,omitempty"`
	Default     *string             `json:"default"`
	Values      *LogStatementValues `json:"values,omitempty"`
}

// LogStatementValues represents the values for the log statement.
type LogStatementValues struct {
	None *string `json:"none,omitempty"`
	DDL  *string `json:"ddl,omitempty"`
	Mod  *string `json:"mod,omitempty"`
	All  *string `json:"all,omitempty"`
}

// SettingsRequest represents a request to update one or more settings on a postgres database.
type SettingsRequest struct {
	// LogLockWaits log when a session waits longer than 1 second to acquire a lock.
	LogLockWaits *bool `json:"log_lock_waits,omitempty"`

	// LogConnections controls whether a log message is produced when a login attempt is made.
	LogConnections *bool `json:"log_connections,omitempty"`

	// LogMinDurationStatement is the duration of each completed statement will be logged if the statement completes
	// after the time specified by a value. This value needs to specified as a whole number, in milliseconds.
	// A value of `0` logs all queries and `-1` disables logging.
	LogMinDurationStatement int `json:"log_min_duration_statement,omitempty"`

	// LogStatement defines which statements are logged.
	// Valid values are:
	// - none: No statements are logged
	// - ddl: All data definition statements, such as CREATE, ALTER and DROP will be logged
	// - mod: Includes all statements from ddl as well as data-modifying statements such as INSERT, UPDATE, DELETE, TRUNCATE, COPY
	// - all: All statements are logged`,
	LogStatement string `json:"log_statement,omitempty"`
}

// GetSettings returns all settings for a postgres database.
func (p *Postgres) GetSettings(nameOrID string) (*Settings, *simpleresty.Response, error) {
	var result *Settings
	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/config", nameOrID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// UpdateSettings updates one or more settings for a postgres database.
//
// NOTE: A successful update request does not necessarily mean the update has been fully applied in Heroku.
// Often times, subsequent requests may return with a 422 status code that indicates the following:
// "Still applying previous configuration change to this database. Please try again later."
func (p *Postgres) UpdateSettings(nameOrID string, opts *SettingsRequest) (*Settings, *simpleresty.Response, error) {
	var result *Settings

	urlStr := p.http.RequestURL("/postgres/v0/databases/%s/config", nameOrID)

	// Execute the request
	response, updateErr := p.http.Patch(urlStr, &result, opts)

	return result, response, updateErr
}
