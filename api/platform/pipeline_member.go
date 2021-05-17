package platform

import (
	"fmt"
	"github.com/davidji99/simpleresty"
	"time"
)

// PipelineMembership represents a Heroku user's membership on a pipeline.
type PipelineMembership struct {
	ID          *string       `json:"id"`
	Pipeline    *Pipeline     `json:"pipeline"`
	User        *User         `json:"user"`
	Permissions []*Permission `json:"permissions"`
	CreatedAt   *time.Time    `json:"created_at" url:"created_at,key"` // when process type was created
	UpdatedAt   *time.Time    `json:"updated_at" url:"updated_at,key"` // when dyno type was updated
}

// PipelineMembershipRequestOpts represents a request to add a member to a pipeline.
type PipelineMembershipRequestOpts struct {
	Permissions []string `json:"permissions"`
	Email       string   `json:"user"`
	PipelineID  string   `json:"pipeline"`
}

// ListPipelineMembers returns all members added to a pipeline.
func (p *Platform) ListPipelineMembers(pipelineID string) ([]*PipelineMembership, *simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	var result []*PipelineMembership

	urlStr := p.http.RequestURL("/pipelines/%s/ephemeral-app-collaborators", pipelineID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", PipelineCollaboratorsAcceptHeader)

	// Execute the request
	response, addErr := p.http.Get(urlStr, &result, nil)

	return result, response, addErr
}

// FindPipelineMembersByEmail retrieves a membership to a pipeline by email.
//
// Returns a PermissionNotFoundError if specified user has not been added to the pipeline.
func (p *Platform) FindPipelineMembersByEmail(pipelineID, email string) (*PipelineMembership, *simpleresty.Response, error) {
	members, listResponse, listErr := p.ListPipelineMembers(pipelineID)
	if listErr != nil {
		return nil, listResponse, listErr
	}

	for _, m := range members {
		if m.GetUser().GetEmail() == email {
			return m, nil, nil
		}
	}

	return nil, nil, PermissionNotFoundError{error: fmt.Errorf("did not find %s on pipeline %s", email, pipelineID)}
}

// AddPipelineMember adds a member to a pipeline.
func (p *Platform) AddPipelineMember(opts *PipelineMembershipRequestOpts) (*PipelineMembership, *simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	var result PipelineMembership

	urlStr := p.http.RequestURL("/ephemeral-app-collaborators")

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", PipelineCollaboratorsAcceptHeader)

	// Execute the request
	response, addErr := p.http.Post(urlStr, &result, opts)

	return &result, response, addErr
}

// UpdatePipelineMemberPermissions modifies a pipeline member's permissions.
func (p *Platform) UpdatePipelineMemberPermissions(membershipID string, permissions []string) (*PipelineMembership, *simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	var result PipelineMembership

	urlStr := p.http.RequestURL("/ephemeral-app-collaborators/%s", membershipID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", PipelineCollaboratorsAcceptHeader)

	opts := struct {
		Permissions []string `json:"permissions"`
	}{
		Permissions: permissions,
	}

	// Execute the request
	response, addErr := p.http.Patch(urlStr, &result, opts)

	return &result, response, addErr
}

// RemovePipelineMember remove a member to a pipeline.
func (p *Platform) RemovePipelineMember(membershipID string) (*simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	urlStr := p.http.RequestURL("/ephemeral-app-collaborators/%s", membershipID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", PipelineCollaboratorsAcceptHeader)

	// Execute the request
	response, addErr := p.http.Delete(urlStr, nil, nil)

	return response, addErr
}
