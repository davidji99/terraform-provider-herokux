package data

import (
	"github.com/davidji99/simpleresty"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/graphql"
)

const (
	pgCredentialSetPermission = `mutation SetPostgresPermissions($addonUUID: ID!, $role: String!, $acls: [PostgresACLInput]!) {
	setPostgresPermissions(
		addon_uuid: $addonUUID
		role: $role
		acls: $acls
	)
}
`
)

type pgSetPermResponse struct {
	SetPostgresPermissions bool `json:"setPostgresPermissions"`
	Extensions map[string]interface{} `json:"extensions"`
}

type acl struct {
	Kind       string
	Name       string
	Default    bool
	Privileges []string
}

func generateACLNoPerms(databaseName string) []acl {
	acls := make([]acl, 0)
	acls = append(acls, acl{
		Kind:       "database",
		Name:       databaseName,
		Privileges: []string{"CONNECT"},
	})
	acls = append(acls, acl{
		Kind:       "table",
		Default:    true,
		Privileges: make([]string, 0),
	})
	acls = append(acls, acl{
		Kind:       "sequence",
		Default:    true,
		Privileges: make([]string, 0),
	})
	acls = append(acls, acl{
		Kind:       "schema",
		Name:       "public",
		Privileges: make([]string, 0),
	})
	acls = append(acls, acl{
		Kind:       "table",
		Name:       "public",
		Privileges: make([]string, 0),
	})
	acls = append(acls, acl{
		Kind:       "sequence",
		Name:       "public",
		Privileges: make([]string, 0),
	})
	return acls
}

func generateACLReadWrite(databaseName string) []acl {
	acls := make([]acl, 0)
	acls = append(acls, acl{
		Kind:       "database",
		Name:       databaseName,
		Privileges: []string{"CONNECT", "TEMPORARY"},
	})
	acls = append(acls, acl{
		Kind:       "table",
		Default:    true,
		Privileges: []string{"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE"},
	})
	acls = append(acls, acl{
		Kind:       "sequence",
		Default:    true,
		Privileges: []string{"SELECT", "USAGE"},
	})
	acls = append(acls, acl{
		Kind:       "schema",
		Name:       "public",
		Privileges: []string{"USAGE"},
	})
	acls = append(acls, acl{
		Kind:       "table",
		Name:       "public",
		Privileges: []string{"SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE"},
	})
	acls = append(acls, acl{
		Kind:       "sequence",
		Name:       "public",
		Privileges: []string{"SELECT", "USAGE"},
	})
	return acls
}

func generateACLReadOnly(databaseName string) []acl {
	acls := make([]acl, 0)
	acls = append(acls, acl{
		Kind:       "database",
		Name:       databaseName,
		Privileges: []string{"CONNECT"},
	})
	acls = append(acls, acl{
		Kind:       "table",
		Default:    true,
		Privileges: []string{"SELECT"},
	})
	acls = append(acls, acl{
		Kind:       "sequence",
		Default:    true,
		Privileges: []string{"SELECT"},
	})
	acls = append(acls, acl{
		Kind:       "schema",
		Name:       "public",
		Privileges: []string{"USAGE"},
	})
	acls = append(acls, acl{
		Kind:       "table",
		Name:       "public",
		Privileges: []string{"SELECT"},
	})
	acls = append(acls, acl{
		Kind:       "sequence",
		Name:       "public",
		Privileges: []string{"SELECT"},
	})
	return acls
}

// SetPGCredentialPermission modifies the permission for a Heroku postgres credential.
func (d *Data) SetPGCredentialPermission(addonID, databaseName, role, permission string) (bool, *simpleresty.Response, error) {
	// This method exists in the Data package instead of the Postgres package as the underlying API
	// uses the Data API base endpoint.

	// Determine which ACLs to use
	var acls []acl
	switch permission {
	case "none":
		acls = generateACLNoPerms(databaseName)
	case "read-only":
		acls = generateACLReadOnly(databaseName)
	case "read-write":
		acls = generateACLReadWrite(databaseName)
	}

	vars := map[string]interface{}{
		"addon_uuid": addonID,
		"role":       role,
		"acls":       acls,
	}

	reqBody := &graphql.Request{
		Query: pgCredentialSetPermission,
		Variables: vars,
	}

	resp := pgSetPermResponse{}
	respBody := &graphql.Response{Data: &resp}

	urlStr := d.http.RequestURL("/graphql")
	response, updateErr := d.http.Post(urlStr, &respBody, reqBody)
	if updateErr != nil {
		return false, response, updateErr
	}

	return resp.SetPostgresPermissions, response, updateErr
}
