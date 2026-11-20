package platform

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/davidji99/simpleresty"
	"github.com/go-resty/resty/v2"
)

// Formation represents an app formation.
type Formation struct {
	App struct {
		ID   *string `json:"id" url:"id,key"`     // unique identifier of app
		Name *string `json:"name" url:"name,key"` // unique name of app
	} `json:"app" url:"app,key"` // app formation belongs to
	Command     *string          `json:"command" url:"command,key"`       // command to use to launch this process
	ID          *string          `json:"id" url:"id,key"`                 // unique identifier of this process type
	Size        *string          `json:"size" url:"size,key"`             // dyno size (default: "standard-1X")
	Type        *string          `json:"type" url:"type,key"`             // type of process to maintain
	Quantity    int              `json:"quantity" url:"quantity,key"`     // number of processes to maintain
	CreatedAt   *time.Time       `json:"created_at" url:"created_at,key"` // when process type was created
	UpdatedAt   *time.Time       `json:"updated_at" url:"updated_at,key"` // when dyno type was updated
	DockerImage *FormationDocker `json:"docker_image,omitempty" url:"docker_image,key"`
}

// FormationDocker represents the information regarding a docker container for a formation.
type FormationDocker struct {
	ID *string `json:"id,omitempty" url:"id,key"`
}

// FormationDockerBatchUpdateOpts represents the batch update opts specifically for releasing docker images.
type FormationDockerBatchUpdateOpts struct {
	Updates []FormationDockerUpdateOpts `json:"updates" url:"updates,key"` // Array with formation updates. Each element must have "type", the id
	// or name of the process type to be updated, and can optionally update
	// its "quantity" or "size".
}

type FormationDockerUpdateOpts struct {
	//Quantity    *int    `json:"quantity,omitempty" url:"quantity,omitempty,key"`         // number of processes to maintain
	//Size        *string `json:"size,omitempty" url:"size,omitempty,key"`                 // dyno size (default: "standard-1X")
	Type          string  `json:"type" url:"type,key"`                           // type of process to maintain
	DockerImageID *string `json:"docker_image" url:"docker_image,omitempty,key"` // algorithm:hex value of the pushed Docker image.
}

// FormationContainerBatchUpdate updates all specified process types with their target container image.
//
// Reference: https://devcenter.heroku.com/articles/container-registry-and-runtime#api
func (p *Platform) FormationContainerBatchUpdate(appIdOrName string, opts *FormationDockerBatchUpdateOpts) (
	[]*Formation, *simpleresty.Response, error) {
	var result []*Formation

	urlStr := p.formation.RequestURL("/apps/%s/formation", appIdOrName)

	// Execute the request
	response, updateErr := p.formation.Patch(urlStr, &result, opts)

	return result, response, updateErr
}

// FormationContainerUpdate updates the specified process type with their target container image.
//
// To destroy an existing process type's container, pass in `nil` for the `docker_image` field in the request body.
func (p *Platform) FormationContainerUpdate(appIdOrName string, processType string, opts *FormationDockerUpdateOpts) (
	*simpleresty.Response, error) {
	urlStr := p.formation.RequestURL("/apps/%s/formation/%s", appIdOrName, processType)

	// Execute the request
	response, updateErr := p.formation.Patch(urlStr, nil, opts)

	return response, updateErr
}

// ReleaseContainerImage only needs the PATCH status. Heroku can omit the formation list.
func (p *Platform) ReleaseContainerImage(ctx context.Context, app, process, image string) error {
	path := p.formation.RequestURL("/apps/%s/formation", url.PathEscape(app))
	opts := FormationDockerBatchUpdateOpts{Updates: []FormationDockerUpdateOpts{{Type: process, DockerImageID: &image}}}
	resp, err := p.formation.R().SetContext(ctx).SetBody(opts).Patch(path)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("releasing %s on %s: %s", process, app, resp.Status())
	}
	return nil
}

func (p *Platform) GetContainerFormation(ctx context.Context, app, process string) (*Formation, *resty.Response, error) {
	path := p.formation.RequestURL("/apps/%s/formation/%s", url.PathEscape(app), url.PathEscape(process))
	var formation Formation
	resp, err := p.formation.R().SetContext(ctx).SetResult(&formation).Get(path)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, resp, fmt.Errorf("read container formation %s: %s", process, resp.Status())
	}
	return &formation, resp, nil
}

func (p *Platform) RemoveContainerFormation(ctx context.Context, app, process, etag string) error {
	if len(etag) < 2 || etag[0] != '"' || etag[len(etag)-1] != '"' {
		return fmt.Errorf("missing strong ETag for container formation %s", process)
	}
	path := p.formation.RequestURL("/apps/%s/formation/%s", url.PathEscape(app), url.PathEscape(process))
	resp, err := p.formation.R().SetContext(ctx).SetHeader("If-Match", etag).SetBody(map[string]any{"docker_image": nil}).Patch(path)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return fmt.Errorf("remove container formation %s: %s", process, resp.Status())
	}
	return nil
}
