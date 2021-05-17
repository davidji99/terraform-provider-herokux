package platform

import (
	"github.com/davidji99/simpleresty"
	"time"
)

// Formation represents an app formation.
type Formation struct {
	App struct {
		ID   *string `json:"id" url:"id,key"`     // unique identifier of app
		Name *string `json:"name" url:"name,key"` // unique name of app
	} `json:"app" url:"app,key"` // app formation belongs to
	Command     *string          `json:"command" url:"command,key"`       // command to use to launch this process
	CreatedAt   *time.Time       `json:"created_at" url:"created_at,key"` // when process type was created
	ID          *string          `json:"id" url:"id,key"`                 // unique identifier of this process type
	Quantity    int              `json:"quantity" url:"quantity,key"`     // number of processes to maintain
	Size        *string          `json:"size" url:"size,key"`             // dyno size (default: "standard-1X")
	Type        *string          `json:"type" url:"type,key"`             // type of process to maintain
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
	defer p.resetAcceptHeader()

	var result []*Formation

	urlStr := p.http.RequestURL("/apps/%s/formation", appIdOrName)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", DockerReleasesAcceptHeader)

	// Execute the request
	response, updateErr := p.http.Patch(urlStr, &result, opts)

	return result, response, updateErr
}

// FormationContainerUpdate updates the specified process type with their target container image.
//
// To destroy an existing process type's container, pass in `nil` for the `docker_image` field in the request body.
func (p *Platform) FormationContainerUpdate(appIdOrName string, processType string, opts *FormationDockerUpdateOpts) (
	*simpleresty.Response, error) {
	defer p.resetAcceptHeader()

	urlStr := p.http.RequestURL("/apps/%s/formation/%s", appIdOrName, processType)

	// TODO: remove custom Accept headers when APIs fully launch.
	p.http.SetHeader("Accept", DockerReleasesAcceptHeader)

	// Execute the request
	response, updateErr := p.http.Patch(urlStr, nil, opts)

	return response, updateErr
}
