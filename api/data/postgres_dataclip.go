package data

import (
	"errors"
	"fmt"
	"github.com/davidji99/simpleresty"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/graphql"
	"github.com/davidji99/terraform-provider-herokux/api/platform"
	"time"
)

// PostgresDataclip represents a postgres data clip.
type PostgresDataclip struct {
	ID           *string                     `json:"id,omitempty"`
	Slug         *string                     `json:"slug,omitempty"`
	Title        *string                     `json:"title,omitempty"`
	TeamID       *string                     `json:"team_id,omitempty"`
	PublicSlug   *string                     `json:"public_slug,omitempty"`
	PublicSlugBy *string                     `json:"public_slug_by,omitempty"` // is the user that enabled sharing
	UserShares   []string                    `json:"user_shares,omitempty"`
	TeamShares   []string                    `json:"team_shares,omitempty"`
	Detached     *bool                       `json:"detached,omitempty"`
	Editable     *bool                       `json:"editable,omitempty"`
	CreatedAt    *time.Time                  `json:"created_at,omitempty"`
	EditedAt     *time.Time                  `json:"edited_at,omitempty"`
	Creator      *platform.User              `json:"creator,omitempty"`
	Datasource   *PostgresDataclipDatasource `json:"datasource,omitempty"`
	Versions     []*PostgresDataclipVersion  `json:"versions,omitempty"`
}

// PostgresDataclipDatasource represents a data clip source.
type PostgresDataclipDatasource struct {
	ID             *string `json:"id,omitempty"`
	AddonID        *string `json:"addon_id,omitempty"`
	AddonName      *string `json:"addon_name,omitempty"`
	AttachmentID   *string `json:"attachment_id,omitempty"`
	AttachmentName *string `json:"attachment_name,omitempty"`
	AppID          *string `json:"app_id,omitempty"`
	AppName        *string `json:"app_name,omitempty"`
}

// PostgresDataclipVersion represents a data clip version.
type PostgresDataclipVersion struct {
	ID               *string                        `json:"id,omitempty"`
	Sql              *string                        `json:"sql,omitempty"`
	URL              *string                        `json:"url,omitempty"`
	CreatorID        *string                        `json:"creator_id,omitempty"`
	Creator          *platform.User                 `json:"creator,omitempty"`
	LatestResultSize *int                           `json:"latest_result_size,omitempty"`
	CreatedAt        *time.Time                     `json:"created_at,omitempty"`
	Result           *PostgresDataclipVersionResult `json:"result,omitempty"`
}

// PostgresDataclipVersionResult represents a data clip query result.
type PostgresDataclipVersionResult struct {
	ID            *string    `json:"id,omitempty"`
	Error         *string    `json:"error,omitempty"`
	QueryStartAt  *time.Time `json:"query_started_at,omitempty"`
	QueryFinishAt *time.Time `json:"query_finished_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	Duration      *int       `json:"duration,omitempty"`
}

type postgresDataclipsListResponse struct {
	ListClips []*PostgresDataclip `json:"listClips"`
}

// ListPostgresDataclips returns all dataclips that the authenticated user has access to in Heroku.
func (d *Data) ListPostgresDataclips() ([]*PostgresDataclip, *simpleresty.Response, error) {
	resp := postgresDataclipsListResponse{}
	respBody := &graphql.Response{Data: &resp}

	urlStr, queryErr := d.http.RequestURLWithQueryParams("/graphql",
		graphql.GetQueryParam{Query: postgresDataclipListKey})
	if queryErr != nil {
		return nil, nil, queryErr
	}

	response, getErr := d.http.Get(urlStr, &respBody, nil)
	if getErr != nil {
		return nil, response, getErr
	}

	if resp.ListClips == nil {
		return nil, nil, errors.New(response.Body)
	}

	return resp.ListClips, response, nil
}

type postgresDataclipGetResponse struct {
	Clip *PostgresDataclip `json:"clip"`
}

// GetPostgresDataclip returns a single dataclip.
func (d *Data) GetPostgresDataclip(slug string) (*PostgresDataclip, *simpleresty.Response, error) {
	resp := postgresDataclipGetResponse{}
	respBody := &graphql.Response{Data: &resp}

	urlStr, queryErr := d.http.RequestURLWithQueryParams("/graphql",
		graphql.GetQueryParam{Query: postgresDataclipGetKey, Variables: fmt.Sprintf("{\"slug\": \"%s\"}", slug)})
	if queryErr != nil {
		return nil, nil, queryErr
	}

	response, getErr := d.http.Get(urlStr, &respBody, nil)
	if getErr != nil {
		return nil, response, getErr
	}

	if resp.Clip == nil {
		return nil, nil, errors.New(response.Body)
	}

	return resp.Clip, response, nil
}

// PostgresDataclipCreateRequest represents a request to create a postgres dataclip.
//
// All fields are required.
type PostgresDataclipCreateRequest struct {
	AttachmentID string
	Sql          string
	Title        string
}

type postgresDataclipCreateResponse struct {
	CreateClip *PostgresDataclip `json:"createClip"`
}

// CreatePostgresDataclip creates a new data clip.
func (d *Data) CreatePostgresDataclip(opts *PostgresDataclipCreateRequest) (*PostgresDataclip, *simpleresty.Response, error) {
	vars := map[string]interface{}{
		"attachmentId": opts.AttachmentID,
		"sql":          opts.Sql,
		"title":        opts.Title,
	}

	reqBody := &graphql.Request{
		Query:     postgresDataclipCreateKey,
		Variables: vars,
	}

	resp := postgresDataclipCreateResponse{}
	respBody := &graphql.Response{Data: &resp}

	urlStr := d.http.RequestURL("/graphql")
	response, createErr := d.http.Post(urlStr, &respBody, reqBody)
	if createErr != nil {
		return nil, response, createErr
	}

	if resp.CreateClip == nil {
		return nil, nil, errors.New(response.Body)
	}

	return resp.CreateClip, response, nil
}

// PostgresDataclipUpdateRequest represents a request to update a Postgres dataclip.
//
// All fields are required.
type PostgresDataclipUpdateRequest struct {
	ClipID string
	PostgresDataclipCreateRequest
}

type postgresDataclipUpdateResponse struct {
	UpdateClip *PostgresDataclip `json:"updateClip"`
}

func (d *Data) UpdatePostgresDataclip(opts *PostgresDataclipUpdateRequest) (*PostgresDataclip, *simpleresty.Response, error) {
	vars := map[string]interface{}{
		"attachmentId": opts.AttachmentID,
		"sql":          opts.Sql,
		"title":        opts.Title,
		"clipId":       opts.ClipID,
	}

	reqBody := &graphql.Request{
		Query:     postgresDataclipUpdateKey,
		Variables: vars,
	}

	resp := postgresDataclipUpdateResponse{}
	respBody := &graphql.Response{Data: &resp}

	urlStr := d.http.RequestURL("/graphql")
	response, createErr := d.http.Post(urlStr, &respBody, reqBody)
	if createErr != nil {
		return nil, response, createErr
	}

	if resp.UpdateClip == nil {
		return nil, nil, errors.New(response.Body)
	}

	return resp.UpdateClip, response, nil
}

type PostgresDataclipDeleteResponse struct {
	DeleteClip string `json:"deleteClip"`
}

func (d *Data) DeleteDataclip(id string) (*PostgresDataclipDeleteResponse, *simpleresty.Response, error) {
	vars := map[string]interface{}{
		"clipId": id,
	}

	reqBody := &graphql.Request{
		Query:     postgresDataclipDeleteKey,
		Variables: vars,
	}

	resp := PostgresDataclipDeleteResponse{}
	respBody := &graphql.Response{Data: &resp}

	urlStr := d.http.RequestURL("/graphql")
	response, deleteErr := d.http.Post(urlStr, &respBody, reqBody)
	if deleteErr != nil {
		return nil, response, deleteErr
	}

	if resp.DeleteClip == "" {
		return nil, nil, errors.New(response.Body)
	}

	return &resp, response, nil
}

type PostgresDataclipSharingResponse struct {
	TogglePublicClipShare *PostgresDataclip `json:"togglePublicClipShare"`
}

func (d *Data) TogglePostgresDataclipSharing(slug string, enabled bool) (*PostgresDataclip, *simpleresty.Response, error) {
	vars := map[string]interface{}{
		"slug": slug,
	}

	query := postgresDataclipDisableShareKey
	if enabled {
		query = postgresDataclipEnableShareKey
	}

	reqBody := &graphql.Request{
		Query:     query,
		Variables: vars,
	}

	resp := PostgresDataclipSharingResponse{}
	respBody := &graphql.Response{Data: &resp}

	urlStr := d.http.RequestURL("/graphql")
	response, deleteErr := d.http.Post(urlStr, &respBody, reqBody)
	if deleteErr != nil {
		return nil, response, deleteErr
	}

	if resp.TogglePublicClipShare == nil {
		return nil, nil, errors.New(response.Body)
	}

	return resp.TogglePublicClipShare, response, nil
}

// TODO: method for https://data-api.heroku.com/dataclips/<DATACLIP_SLUG>.json

const (
	postgresDataclipListKey = `
query ListClips {
    listClips {
      id
      slug
      title
      creator {
        id
        email
      }
      created_at
      edited_at
      datasource {
        addon_name
        attachment_id
      }
      detached
      editable
    }
  }
`

	postgresDataclipGetKey = `
query FetchClipDetails($slug: ID!) {
    clip(slug: $slug) {
      ...clipFragment
    }
  }` + graphqlAPIPostgresDataclipFields

	postgresDataclipDeleteKey = `
mutation DeleteDataclip($clipId: ID!) {
    deleteClip(clipId: $clipId)
}
`

	postgresDataclipCreateKey = `
mutation CreatePostgresDataclip($attachmentId: ID!, $title: String!, $sql: String!, $teamId: ID) {
    createClip(attachmentId: $attachmentId, title: $title, sql: $sql, teamId: $teamId) {
        ...clipFragment   
        }  
    }` + graphqlAPIPostgresDataclipFields

	postgresDataclipUpdateKey = `
mutation UpdateDataclip($clipId: ID!, $attachmentId: ID!, $title: String!, $sql: String!) {
    updateClip(clipId: $clipId, attachmentId: $attachmentId, title: $title, sql: $sql) {
        ...clipFragment
        }
    }` + graphqlAPIPostgresDataclipFields

	postgresDataclipEnableShareKey = `
mutation SharePublicDataclip($slug: ID!) {
    togglePublicClipShare(slug: $slug, enabled: true) {
        ...clipFragment
    }
}` + graphqlAPIPostgresDataclipFields

	postgresDataclipDisableShareKey = `
mutation UnsharePublicDataclip($slug: ID!) {
    togglePublicClipShare(slug: $slug, enabled: false) {
        ...clipFragment
    }
}` + graphqlAPIPostgresDataclipFields

	graphqlAPIPostgresDataclipFields = `
fragment clipFragment on Clip {
        id
        created_at
        creator {
            id
            email
        }
        edited_at
        slug
        title
        user_shares {
            id
            clip_id
            shared_by {
                id
                email
            }
            shared_with {
                id
                email
            }
        }
        team_shares {
            id
            clip_id
            shared_by {
                id
                email
            }
            shared_with {
                id
                name
            }
        }
        team_id
        public_slug
        public_slug_by
        detached
        datasource {
            id
            addon_id
            addon_name
            attachment_id
            attachment_name
            app_id
            app_name
        }
        versions(limit: 1) {
            id
            created_at
            sql
            url
            latest_result_checksum
            latest_result_at
            latest_result_size
            creator_id
            creator {
                email
            }
            result {
                id
                query_started_at
                query_finished_at
                error
                completed_at
                duration
            }
        }
        editable
}
`
)
