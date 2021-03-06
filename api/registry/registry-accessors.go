// Copyright 2020
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Code generated by gen-accessors; DO NOT EDIT.
package registry

// GetConfig returns the Config field.
func (m *Manifest) GetConfig() *ManifestConfig {
	if m == nil {
		return nil
	}
	return m.Config
}

// HasLayers checks if Manifest has any Layers.
func (m *Manifest) HasLayers() bool {
	if m == nil || m.Layers == nil {
		return false
	}
	if len(m.Layers) == 0 {
		return false
	}
	return true
}

// GetMediaType returns the MediaType field if it's non-nil, zero value otherwise.
func (m *Manifest) GetMediaType() string {
	if m == nil || m.MediaType == nil {
		return ""
	}
	return *m.MediaType
}

// GetSchemaVersion returns the SchemaVersion field if it's non-nil, zero value otherwise.
func (m *Manifest) GetSchemaVersion() int {
	if m == nil || m.SchemaVersion == nil {
		return 0
	}
	return *m.SchemaVersion
}

// GetDigest returns the Digest field if it's non-nil, zero value otherwise.
func (m *ManifestConfig) GetDigest() string {
	if m == nil || m.Digest == nil {
		return ""
	}
	return *m.Digest
}

// GetMediaType returns the MediaType field if it's non-nil, zero value otherwise.
func (m *ManifestConfig) GetMediaType() string {
	if m == nil || m.MediaType == nil {
		return ""
	}
	return *m.MediaType
}

// GetSize returns the Size field if it's non-nil, zero value otherwise.
func (m *ManifestConfig) GetSize() int {
	if m == nil || m.Size == nil {
		return 0
	}
	return *m.Size
}

// GetDigest returns the Digest field if it's non-nil, zero value otherwise.
func (m *ManifestLayer) GetDigest() string {
	if m == nil || m.Digest == nil {
		return ""
	}
	return *m.Digest
}

// GetMediaType returns the MediaType field if it's non-nil, zero value otherwise.
func (m *ManifestLayer) GetMediaType() string {
	if m == nil || m.MediaType == nil {
		return ""
	}
	return *m.MediaType
}

// GetSize returns the Size field if it's non-nil, zero value otherwise.
func (m *ManifestLayer) GetSize() int {
	if m == nil || m.Size == nil {
		return 0
	}
	return *m.Size
}
