package postgres

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	"time"
)

// DataLink represents a data link between two postgres databases.
type DataLink struct {
	ID         *string         `json:"id,omitempty"`
	CreatedAt  *time.Time      `json:"created_at,omitempty"`
	RemoteName *string         `json:"remote_name,omitempty"`
	Remote     *DataLinkRemote `json:"remote,omitempty"`

	// Name of the data link. If no name is defined, this value is the same as the RemoteName.
	Name *string `json:"name,omitempty"`
}

// DataLinkRemote represents the remote link.
type DataLinkRemote struct {
	Name           *string `json:"name,omitempty"`
	AttachmentName *string `json:"attachment_name,omitempty"`
}

// DataLinkCreateOpts represents a request to create a data link.
type DataLinkCreateOpts struct {
	// Remote - The data store that is being connected to a Heroku Postgres database.
	Remote string `json:"target,omitempty"`

	// Name - The name of connection between the remote and local databases.
	Name string `json:"as,omitempty"`
}

// ListDataLink lists all data links for a postgres database.
func (p *Postgres) ListDataLink(localDbID string) ([]*DataLink, *simpleresty.Response, error) {
	var result []*DataLink
	urlStr := p.http.RequestURL("/client/v11/databases/%s/links", localDbID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// FindDataLinkByID finds a data link by its ID.
func (p *Postgres) FindDataLinkByID(localDbID, dataLinkID string) (*DataLink, *simpleresty.Response, error) {
	links, listResponse, listErr := p.ListDataLink(localDbID)
	if listErr != nil {
		return nil, listResponse, listErr
	}

	for _, l := range links {
		if l.GetID() == dataLinkID {
			return l, nil, nil
		}
	}

	return nil, nil, fmt.Errorf("did not find data link by ID %s", dataLinkID)
}

// FindDataLinkByName finds a data link by its name.
func (p *Postgres) FindDataLinkByName(localDbID, dataLinkName string) (*DataLink, *simpleresty.Response, error) {
	links, listResponse, listErr := p.ListDataLink(localDbID)
	if listErr != nil {
		return nil, listResponse, listErr
	}

	for _, l := range links {
		if l.GetName() == dataLinkName {
			return l, nil, nil
		}
	}

	return nil, nil, fmt.Errorf("did not find data link by name %s", dataLinkName)
}

// CreateDataLink creates a data link between two databases.
func (p *Postgres) CreateDataLink(localDbID string, opts *DataLinkCreateOpts) (*DataLink, *simpleresty.Response, error) {
	var result *DataLink
	urlStr := p.http.RequestURL("/client/v11/databases/%s/links", localDbID)

	// Execute the request
	response, getErr := p.http.Post(urlStr, &result, opts)

	return result, response, getErr
}

// DeleteDataLink deletes a data link between two databases.
func (p *Postgres) DeleteDataLink(localDbID, linkName string) (*simpleresty.Response, error) {
	urlStr := p.http.RequestURL("/client/v11/databases/%s/links/%s", localDbID, linkName)

	// Execute the request
	response, getErr := p.http.Delete(urlStr, nil, nil)

	return response, getErr
}
