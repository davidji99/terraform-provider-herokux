package kolkrabbi

import (
	"github.com/davidji99/simpleresty"
	"github.com/davidji99/terraform-provider-herokux/api/platform"
	"time"
)

// AppGHIntegration represents the integration between a Heroku app and Github repository.
type AppGHIntegration struct {
	App        *platform.App              `json:"organization,omitempty"`
	Owner      *PipelineGHIntegrationUser `json:"owner,omitempty"`
	AutoDeploy *bool                      `json:"auto_deploy,omitempty"`
	WaitForCI  *bool                      `json:"wait_for_ci,omitempty"`
	AppID      *string                    `json:"app_id,omitempty"`
	ID         *string                    `json:"id,omitempty"`
	Branch     *string                    `json:"branch,omitempty"`
	Repo       *string                    `json:"repo,omitempty"`
	RepoID     *int                       `json:"repo_id,omitempty"`
	CreatedAt  *time.Time                 `json:"created_at,omitempty"`
	UpdatedAt  *time.Time                 `json:"updated_at,omitempty"`
	StaleDays  *int                       `json:"stale_days,omitempty"`
}

// AppGhIntegrationRequest represents a request to modify the integration.
type AppGhIntegrationRequest struct {
	AutoDeploy *bool  `json:"auto_deploy,omitempty"`
	Branch     string `json:"branch,omitempty"`
	WaitForCI  *bool  `json:"wait_for_ci,omitempty"`
}

// GetAppGithubIntegration returns information regarding the integration between a Heroku app and Github repository.
func (k *Kolkrabbi) GetAppGithubIntegration(appID string) (*AppGHIntegration, *simpleresty.Response, error) {
	var result AppGHIntegration
	urlStr := k.http.RequestURL("/apps/%s/github", appID)

	// Execute the request
	response, getErr := k.http.Get(urlStr, &result, nil)

	return &result, response, getErr
}

// UpdateAppGithubIntegration updates the integration between a Heroku app and Github repository.
func (k *Kolkrabbi) UpdateAppGithubIntegration(appID string, opts *AppGhIntegrationRequest) (*AppGHIntegration, *simpleresty.Response, error) {
	var result AppGHIntegration
	urlStr := k.http.RequestURL("/apps/%s/github", appID)

	// Execute the request
	response, getErr := k.http.Patch(urlStr, &result, opts)

	return &result, response, getErr
}
