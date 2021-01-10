package registry

import "github.com/davidji99/simpleresty"

type Manifest struct {
	SchemaVersion *int             `json:"schemaVersion,omitempty"`
	MediaType     *string          `json:"mediaType,omitempty"`
	Config        *ManifestConfig  `json:"config,omitempty"`
	Layers        []*ManifestLayer `json:"layers,omitempty"`
}

type ManifestConfig struct {
	MediaType *string `json:"mediaType,omitempty"`
	Size      *int    `json:"size,omitempty"`

	// This digest represents the docker image ID that Heroku wants
	// as documented here https://devcenter.heroku.com/articles/container-registry-and-runtime#getting-a-docker-image-id.
	Digest *string `json:"digest,omitempty"`
}

type ManifestLayer struct {
	MediaType *string `json:"mediaType,omitempty"`
	Size      *int    `json:"size,omitempty"`
	Digest    *string `json:"digest,omitempty"`
}

// ListAppProcessManifests lists pushed docker images by tag.
//
// The tag parameter is usually 'latest'.
func (r *Registry) ListAppProcessManifests(appIDorName, processType, tag string) (*Manifest, *simpleresty.Response, error) {
	var result *Manifest

	urlStr := r.http.RequestURL("/v2/%s/%s/manifests/%s", appIDorName, processType, tag)

	// Execute the request
	response, getErr := r.http.Get(urlStr, &result, nil)

	return result, response, getErr
}
