package postgres

import (
	"encoding/json"
	"github.com/davidji99/simpleresty"
)

// BackupSchedule represents a postgres backup schedule.
type BackupSchedule struct {
	ID           *string      `json:"uuid,omitempty"`
	Name         *string      `json:"name,omitempty"`
	Hour         *json.Number `json:"hour,omitempty"`
	Days         []string     `json:"days,omitempty"`
	Timezone     *string      `json:"timezone,omitempty"`
	CreatedAt    *string      `json:"created_at,omitempty"`
	UpdatedAt    *string      `json:"updated_at,omitempty"`
	DeletedAt    *string      `json:"deleted_at,omitempty"`
	RetainWeeks  *int         `json:"retain_weeks,omitempty"`
	RetainMonths *int         `json:"retain_months,omitempty"`
}

// BackupScheduleRequest represents a request to create/modify a backup schedule.
type BackupScheduleRequest struct {
	// Hour can be int or string but will use string for now. Valid options are 0-23.
	Hour string `json:"hour,omitempty"`

	// Timezone uses https://en.wikipedia.org/wiki/List_of_tz_database_time_zones.
	Timezone string `json:"timezone,omitempty"`
}

// ListBackupSchedules returns all backup schedules for a database.
func (p *Postgres) ListBackupSchedules(nameOrID string) ([]*BackupSchedule, *simpleresty.Response, error) {
	var result []*BackupSchedule
	urlStr := p.http.RequestURL("/client/v11/databases/%s/transfer-schedules", nameOrID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// CreateBackupSchedule creates a backup schedule.
//
// This method effectively acts as a PUT request as well.
func (p *Postgres) CreateBackupSchedule(nameOrID string, opts *BackupScheduleRequest) (*BackupSchedule, *simpleresty.Response, error) {
	var result *BackupSchedule
	urlStr := p.http.RequestURL("/client/v11/databases/%s/transfer-schedules", nameOrID)

	// Execute the request
	response, createErr := p.http.Post(urlStr, &result, opts)

	return result, response, createErr
}

// DeleteBackupSchedule deletes a backup schedule.
func (p *Postgres) DeleteBackupSchedule(dbNameOrID, scheduleID string) (*simpleresty.Response, error) {
	urlStr := p.http.RequestURL("/client/v11/databases/%s/transfer-schedules/%s", dbNameOrID, scheduleID)

	// Execute the request
	response, getErr := p.http.Delete(urlStr, nil, nil)

	return response, getErr
}
