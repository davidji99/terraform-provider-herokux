package data

import (
	"github.com/davidji99/simpleresty"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/graphql"
)

// Privatelink represents a connection between a Heroku postgres and AWS resources.
type Privatelink struct {
	AppName         *string                       `json:"app_name,omitempty"`
	AddonName       *string                       `json:"addon_name,omitempty"`
	AddonUUID       *string                       `json:"addon_uuid,omitempty"`
	Status          *string                       `json:"status,omitempty"`
	ServiceName     *string                       `json:"service_name,omitempty"`
	AllowedAccounts []*PrivatelinkAllowedAccounts `json:"allowed_accounts,omitempty"`
	Connections     []*PrivatelinkConnections     `json:"connections,omitempty"`
}

// PrivatelinkAllowedAccounts represents AWS accounts granted access to a privatelink.
type PrivatelinkAllowedAccounts struct {
	ARN    *string `json:"arn,omitempty"`
	Status *string `json:"status,omitempty"`
}

// PrivatelinkConnections represents connections between Heroku postgres and AWS resources.
type PrivatelinkConnections struct {
	EndpointID *string `json:"endpoint_id,omitempty"`
	Hostname   *string `json:"hostname,omitempty"`
	OwnerARN   *string `json:"owner_arn,omitempty"`
	Status     *string `json:"status,omitempty"`
	graphql.ErrorResponse
}

type privatelinkGetResponse struct {
	Privatelink Privatelink `json:"privatelink"`
}

func (d *Data) GetPrivatelink(addonID string) (*Privatelink, *simpleresty.Response, error) {
	vars := map[string]interface{}{
		"addonUUID": addonID,
	}

	reqBody := &graphql.Request{
		Query:     privatelinkGetKey,
		Variables: vars,
	}

	resp := privatelinkGetResponse{}
	respBody := &graphql.Response{Data: &resp}

	urlStr := d.http.RequestURL("/graphql")
	response, getErr := d.http.Post(urlStr, &respBody, reqBody)
	if getErr != nil {
		return nil, response, getErr
	}

	return &resp.Privatelink, response, nil
}

const (
	privatelinkGetKey = `
query FetchPrivatelink($addonUUID: ID!) {
    privatelink(addonUUID: $addonUUID) {
      ...privatelinkFragment
    }
  }fragment privatelinkFragment on Privatelink {
  app_name
  addon_name
  addon_uuid
  status
  service_name
  allowed_accounts {
    arn
    status
  }
  connections {
    endpoint_id
    hostname
    owner_arn
    status
  }
}
`
)
