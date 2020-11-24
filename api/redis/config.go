package redis

import "github.com/davidji99/simpleresty"

const (
	DisableNotifyKeyspaceEvents = "disabled"
)

// Config represents the configuration options for a Redis instance.
type Config struct {
	MaxmemoryPolicy       *ConfigMaxmemory             `json:"maxmemory_policy,omitempty"`
	NotifyKeyspaceEvents  *ConfigNotifyKeyspaceEvents  `json:"notify_keyspace_events,omitempty"`
	Timeout               *ConfigTimeout               `json:"timeout,omitempty"`
	StandbySegvWorkaround *ConfigStandbySegvWorkaround `json:"standby_segv_workaround,omitempty"`
}

// ConfigMaxmemory represents the max memory configuration.
type ConfigMaxmemory struct {
	Value       *string `json:"value,omitempty"`
	Description *string `json:"desc,omitempty"`
	Default     *string `json:"default,omitempty"`
}

// ConfigNotifyKeyspaceEvents represents the keyspace event notification configuration.
type ConfigNotifyKeyspaceEvents struct {
	Value       *string `json:"value,omitempty"`
	Description *string `json:"desc,omitempty"`
	Default     *string `json:"default,omitempty"`
}

// ConfigMaxmemory represents the timeout configuration.
type ConfigTimeout struct {
	Value       *int    `json:"value,omitempty"`
	Description *string `json:"desc,omitempty"`
	Default     *int    `json:"default,omitempty"`
}

type ConfigStandbySegvWorkaround struct {
	Value       *bool   `json:"value,omitempty"`
	Description *string `json:"desc,omitempty"`
	Default     *bool   `json:"default,omitempty"`
}

// ConfigUpdateRequest represents a request to update Redis configuration(s).
type ConfigUpdateRequest struct {
	// Sets the desired key eviction behavior when a Redis instance reaches max memory.
	MaxmemoryPolicy string `json:"maxmemory_policy,omitempty"`

	// Enables keyspace notifications.
	NotifyKeyspaceEvents *string `json:"notify_keyspace_events,omitempty"`

	// Represent how long a client can idle before being disconnected.
	Timeout *int `json:"timeout,omitempty"`
}

// GetConfig retrieves configurations for a Redis instance.
func (r *Redis) GetConfig(id string) (*Config, *simpleresty.Response, error) {
	var result *Config
	urlStr := r.http.RequestURL("/redis/v0/databases/%s/config", id)

	// Execute the request
	response, getErr := r.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// UpdateConfig updates Redis configuration(s).
func (r *Redis) UpdateConfig(id string, opts *ConfigUpdateRequest) (*Config, *simpleresty.Response, error) {
	var result *Config
	urlStr := r.http.RequestURL("/redis/v0/databases/%s/config", id)

	// Execute the request
	response, updateErr := r.http.Patch(urlStr, &result, opts)

	return result, response, updateErr
}
