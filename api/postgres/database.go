package postgres

import "github.com/davidji99/simpleresty"

// Database represents a Heroku postgres database
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

// DatabaseLeader represents a database's leader
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

func (p *Postgres) GetDatabase(dbID string) (*Database, *simpleresty.Response, error) {
	var result *Database
	urlStr := p.http.RequestURL("/client/v11/databases/%s", dbID)

	// Execute the request
	response, createErr := p.http.Get(urlStr, &result, nil)

	return result, response, createErr
}

func (d *Database) FindInfoByName(name string) *DatabaseInfo {
	for _, i := range d.Info {
		if i.GetName() == name {
			return i
		}
	}

	return nil
}
