package kolkrabbi

import (
	"github.com/davidji99/simpleresty"
	"time"
)

// Pipeline represents a Heroku pipeline.
type Pipeline struct {
	ID *string `json:"id"`
}

// PipelineGHIntegration represents the integration between a Heroku pipeline and Github.
type PipelineGHIntegration struct {
	CI           *bool                      `json:"ci"`
	ID           *string                    `json:"id"`
	Creator      *PipelineGHIntegrationUser `json:"creator"`
	Owner        *PipelineGHIntegrationUser `json:"owner"`
	Repository   *PipelineRepository        `json:"repository"`
	Pipeline     *Pipeline                  `json:"pipeline"`
	Organization interface{}                `json:"organization,omitempty"`
	CreatedAt    *time.Time                 `json:"created_at"`
	UpdatedAt    *time.Time                 `json:"updated_at"`
}

type PipelineGHIntegrationUser struct {
	ID     *string                          `json:"id"`
	Heroku *PipelineGHIntegrationHerokuUser `json:"heroku"`
	Github *PipelineGHIntegrationGithubUser `json:"github"`
}

type PipelineGHIntegrationHerokuUser struct {
	UserID *string `json:"user_id"`
}

type PipelineGHIntegrationGithubUser struct {
	UserID *int `json:"user_id"`
}

type PipelineRepository struct {
	ID        *int       `json:"id"`
	Name      *string    `json:"name"`
	Type      *string    `json:"type"`
	UpdatedAt *time.Time `json:"updated_at"`
	CreatedAt *time.Time `json:"created_at"`
}

// PipelineGHIntegrationRequest represents the request to create/modify the integration.
type PipelineGHIntegrationRequest struct {
	// Repository represents the Github repository ID.
	Repository int64 `json:"repository"`
}

// GetPipelineGithubIntegration retrieves information about a pipeline's integration with a Github repository.
func (k *Kolkrabbi) GetPipelineGithubIntegration(pipelineID string) (*PipelineGHIntegration, *simpleresty.Response, error) {
	var result PipelineGHIntegration
	urlStr := k.http.RequestURL("/pipelines/%s/repository", pipelineID)

	// Execute the request
	response, getErr := k.http.Get(urlStr, &result, nil)

	return &result, response, getErr
}

// CreatePipelineGithubIntegration creates the integration between a pipeline and Github repository.
func (k *Kolkrabbi) CreatePipelineGithubIntegration(pipelineID string, opts *PipelineGHIntegrationRequest) (*PipelineGHIntegration, *simpleresty.Response, error) {
	var result PipelineGHIntegration
	urlStr := k.http.RequestURL("/pipelines/%s/repository", pipelineID)

	// Execute the request
	response, createErr := k.http.Post(urlStr, &result, opts)

	return &result, response, createErr
}

// DeletePipelineGithubIntegration destroys the integration between a pipeline and Github repository.
func (k *Kolkrabbi) DeletePipelineGithubIntegration(pipelineID string) (*simpleresty.Response, error) {
	urlStr := k.http.RequestURL("/pipelines/%s/repository", pipelineID)

	// Execute the request
	response, deleteErr := k.http.Delete(urlStr, nil, nil)

	return response, deleteErr
}
