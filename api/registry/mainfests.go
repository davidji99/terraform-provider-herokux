package registry

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidji99/simpleresty"
)

// Manifest represents a docker image data. This struct is only returned via a specific header, DistributionManifestAcceptHeader.
type Manifest struct {
	History []struct {
		Compatibility string `json:"v1Compatibility"`
	} `json:"history,omitempty"`
	FSLayers      []struct{}       `json:"fsLayers,omitempty"`
	SchemaVersion *int             `json:"schemaVersion,omitempty"`
	MediaType     *string          `json:"mediaType,omitempty"`
	Config        *ManifestConfig  `json:"config,omitempty"`
	Layers        []*ManifestLayer `json:"layers,omitempty"`
}

// ReleaseImage returns the config digest used by container:release and the layer count.
func (m *Manifest) ReleaseImage() (string, int, error) {
	switch m.GetSchemaVersion() {
	case 1:
		var config struct {
			ID string `json:"id"`
		}
		if len(m.History) == 0 || json.Unmarshal([]byte(m.History[0].Compatibility), &config) != nil || config.ID == "" {
			return "", 0, fmt.Errorf("missing or invalid Docker v1 image history")
		}
		return "sha256:" + strings.TrimPrefix(config.ID, "sha256:"), len(m.FSLayers), nil
	case 2:
		return m.GetConfig().GetDigest(), len(m.Layers), nil
	default:
		return "", 0, fmt.Errorf("unsupported Docker manifest schema version %d", m.GetSchemaVersion())
	}
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
