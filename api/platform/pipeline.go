package platform

import (
	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/simpleresty"
)

// Pipeline represents a Heroku pipeline.
type Pipeline struct {
	// Embed the heroku-go Pipeline struct
	*heroku.Pipeline

	EphemeralApps *PipelinePermissionConfiguration `json:"ephemeral_apps,omitempty"`
}

// PipelinePermissionConfiguration represents the permission configuration for review and CI apps on a pipeline.
type PipelinePermissionConfiguration struct {
	Enabled         *bool         `json:"collaborators_enabled,omitempty"`
	Synchronization *bool         `json:"collaborator_synchronization,omitempty"`
	Permissions     []*Permission `json:"collaborator_permissions,omitempty"`
}

// PipelinePermissionConfigUpdateOpts represents a request to modify a pipeline's permissions.
//
// Notes about the update request:
// - The value of the Enabled field doesn't matter.
// - The value of the Synchronization field controls auto-join.
// - If Enabled is `false` but Synchronization is `true`, auto-join is still enabled.
// - Keep Enabled to `true` at all times.
// - By default, 'view' permission is always set even if it is not present in the request.
// - If you only define 'deploy' for a permission, 'view' is automatically added.
type PipelinePermissionConfigUpdateOpts struct {
	Enabled         bool     `json:"collaborators_enabled"`
	Synchronization bool     `json:"collaborator_synchronization"`
	Permissions     []string `json:"collaborator_permissions,omitempty"`
}

// GetPipelinePermissionConfig returns information about a pipeline's permission configuration.
//
// This method also returns basic information about the pipeline itself.
func (p *Platform) GetPipelinePermissionConfig(pipelineID string) (*Pipeline, *simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	var result Pipeline

	urlStr := p.http.RequestURL("/pipelines/%s", pipelineID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", PipelineCollaboratorsAcceptHeader)

	// Execute the request
	response, updateErr := p.http.Get(urlStr, &result, nil)

	return &result, response, updateErr
}

// UpdatePipelinePermissionConfig updates an existing pipeline permission configuration.
func (p *Platform) UpdatePipelinePermissionConfig(pipelineID string, opts *PipelinePermissionConfigUpdateOpts) (*Pipeline, *simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	var result Pipeline

	urlStr := p.http.RequestURL("/pipelines/%s", pipelineID)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", PipelineCollaboratorsAcceptHeader)

	// Construct request body
	o := struct {
		EphemeralApps *PipelinePermissionConfigUpdateOpts `json:"ephemeral_apps"`
	}{
		EphemeralApps: opts,
	}

	// Execute the request
	response, updateErr := p.http.Patch(urlStr, &result, o)

	return &result, response, updateErr
}
