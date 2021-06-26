package registry

import "github.com/davidji99/simpleresty"

// Manifest represents a docker image data. This struct is only returned via a specific header, DistributionManifestAcceptHeader.
type Manifest struct {
	SchemaVersion *int             `json:"schemaVersion,omitempty"`
	MediaType     *string          `json:"mediaType,omitempty"`
	Config        *ManifestConfig  `json:"config,omitempty"`
	Layers        []*ManifestLayer `json:"layers,omitempty"`
}

// ManifestConfig is a config for a manifest.
type ManifestConfig struct {
	MediaType *string `json:"mediaType,omitempty"`
	Size      *int    `json:"size,omitempty"`

	// This digest represents the docker image ID that Heroku wants
	// as documented here https://devcenter.heroku.com/articles/container-registry-and-runtime#getting-a-docker-image-id.
	Digest *string `json:"digest,omitempty"`
}

// ManifestLayer represents a docker layer.
type ManifestLayer struct {
	MediaType *string `json:"mediaType,omitempty"`
	Size      *int    `json:"size,omitempty"`
	Digest    *string `json:"digest,omitempty"`
}

// GetAppProcessManifests retrieves a pushed docker images by tag.
//
// Note: the only acceptable tag parameter is 'latest' as 01-11-2021.
func (r *Registry) GetAppProcessManifests(appIDorName, processType, tag string) (*Manifest, *simpleresty.Response, error) {
	var result *Manifest

	urlStr := r.http.RequestURL("/v2/%s/%s/manifests/%s", appIDorName, processType, tag)

	// Execute the request
	response, getErr := r.http.Get(urlStr, &result, nil)

	return result, response, getErr
}
