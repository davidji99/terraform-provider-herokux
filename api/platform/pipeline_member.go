package platform

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/davidji99/simpleresty"
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
	return listPages[*PipelineMembership](context.Background(), p.pipelines, p.pipelines.RequestURL("/pipelines/%s/ephemeral-app-collaborators", pipelineID))
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
		if m.GetUser().GetEmail() == "" || m.GetID() == "" || m.GetPipeline() == nil || m.GetPipeline().Pipeline == nil || !strings.EqualFold(m.GetPipeline().ID, pipelineID) {
			return nil, listResponse, fmt.Errorf("invalid membership on pipeline %s", pipelineID)
		}
		if m.GetUser().GetEmail() == email {
			return m, nil, nil
		}
	}

	return nil, nil, PermissionNotFoundError{error: fmt.Errorf("did not find %s on pipeline %s", email, pipelineID)}
}

// AddPipelineMember adds a member to a pipeline.
func (p *Platform) AddPipelineMember(opts *PipelineMembershipRequestOpts) (*PipelineMembership, *simpleresty.Response, error) {
	var result PipelineMembership

	urlStr := p.pipelines.RequestURL("/ephemeral-app-collaborators")

	// Execute the request
	response, addErr := p.pipelines.Post(urlStr, &result, opts)

	return &result, response, addErr
}

// UpdatePipelineMemberPermissions modifies a pipeline member's permissions.
func (p *Platform) UpdatePipelineMemberPermissions(membershipID string, permissions []string) (*PipelineMembership, *simpleresty.Response, error) {
	var result PipelineMembership

	urlStr := p.pipelines.RequestURL("/ephemeral-app-collaborators/%s", membershipID)

	opts := struct {
		Permissions []string `json:"permissions"`
	}{
		Permissions: permissions,
	}

	// Execute the request
	response, addErr := p.pipelines.Patch(urlStr, &result, opts)

	return &result, response, addErr
}

// RemovePipelineMember remove a member to a pipeline.
func (p *Platform) RemovePipelineMember(membershipID string) (*simpleresty.Response, error) {
	urlStr := p.pipelines.RequestURL("/ephemeral-app-collaborators/%s", membershipID)

	// Execute the request
	response, addErr := p.pipelines.Delete(urlStr, nil, nil)

	return response, addErr
}
