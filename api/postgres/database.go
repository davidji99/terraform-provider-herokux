package postgres

import (
	"fmt"
	"github.com/davidji99/simpleresty"
)

// Database represents a Heroku postgres database.
type Database struct {
	Following           *string         `json:"following,omitempty"`
	HotStandby          *bool           `json:"hot_standby,omitempty"`
	AddonID             *string         `json:"addon_id,omitempty"`
	Name                *string         `json:"name,omitempty"`
	HerokuResourceID    *string         `json:"heroku_resource_id,omitempty"`
	MetaasSource        *string         `json:"metaas_source,omitempty"`
	PostgresVersion     *string         `json:"postgres_version,omitempty"`
	AvailableForIngress *bool           `json:"available_for_ingress,omitempty"`
	ResourceURL         *string         `json:"resource_url,omitempty"`
	Waiting             *bool           `json:"waiting?,omitempty"`
	Leader              *DatabaseLeader `json:"leader,omitempty"`
	Info                []*DatabaseInfo `json:"info,omitempty"`
}

// DatabaseLeader represents a database's leader.
type DatabaseLeader struct {
	AddonID *string `json:"addon_id,omitempty"`
	Name    *string `json:"name,omitempty"`
}

// DatabaseInfo represents a database's information.
type DatabaseInfo struct {
	Name          *string       `json:"name,omitempty"`
	Values        []interface{} `json:"values,omitempty"` // most of the values are strings
	ResolveDBName *bool         `json:"resolve_db_name,omitempty"`
}

func (d *Database) RetrieveSpecificInfo(name string) (*DatabaseInfo, error) {
	for _, info := range d.Info {
		if info.GetName() == name {
			return info, nil
		}
	}
	return nil, fmt.Errorf("specific DB info not found")
}

// DatabaseWaitStatus represents the status of a database.
type DatabaseWaitStatus struct {
	Status    *string `json:"message,omitempty"`
	IsWaiting *string `json:"waiting?,omitempty"`
}

// GetDB returns detailed information about a Heroku postgres database.
func (p *Postgres) GetDB(dbID string) (*Database, *simpleresty.Response, error) {
	var result *Database
	urlStr := p.http.RequestURL("/client/v11/databases/%s", dbID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// GetDBWaitStatus returns the database's overall status and whether or not it is waiting.
func (p *Postgres) GetDBWaitStatus(dbID string) (*DatabaseWaitStatus, *simpleresty.Response, error) {
	var result *DatabaseWaitStatus
	urlStr := p.http.RequestURL("/client/v11/databases/%s/wait_status", dbID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// UnfollowDB tells a follower DB to unfollow the leader DB.
func (p *Postgres) UnfollowDB(dbID string) (*GenericResponse, *simpleresty.Response, error) {
	var result *GenericResponse
	urlStr := p.http.RequestURL("/client/v11/databases/%s/unfollow", dbID)

	// Construct request body
	body := struct {
		Host string `json:"host"`
	}{Host: ""}

	// Execute the request
	response, err := p.http.Put(urlStr, &result, &body)

	return result, response, err
}

func (d *Database) FindInfoByName(name string) *DatabaseInfo {
	for _, i := range d.Info {
		if i.GetName() == name {
			return i
		}
	}

	return nil
}
