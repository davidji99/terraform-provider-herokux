package kafka

import (
	"fmt"
	"github.com/davidji99/simpleresty"
)

// Topics represents a list of topics.
type Topics struct {
	AddonAttachmentConfigVar *string      `json:"attachment_name,omitempty"`
	Prefix                   *string      `json:"prefix,omitempty"`
	Topics                   []*Topic     `json:"topics,omitempty"`
	Limits                   *TopicLimits `json:"limits,omitempty"`
}

// Topic represents a single topic.
type Topic struct {
	Name               *string `json:"name,omitempty"`
	Prefix             *string `json:"prefix,omitempty"`
	MessageInPerSecond *int    `json:"messages_in_per_second,omitempty"`
	BytesInPerSecond   *int    `json:"bytes_in_per_second,omitempty"`
	BytesOutPerSecond  *int    `json:"bytes_out_per_second,omitempty"`
	Partitions         *int    `json:"partitions,omitempty"`
	ReplacementFactor  *int    `json:"replication_factor,omitempty"`
	Status             *string `json:"status,omitempty"`
	StatusLabel        *string `json:"status_label,omitempty"`
	DataSize           *int    `json:"data_size,omitempty"`
	Compaction         *bool   `json:"compaction,omitempty"`
	RetentionTimeInMS  *int    `json:"retention_time_ms,omitempty"`
	CleanupPolicy      *string `json:"cleanup_policy,omitempty"`
	CompactionEnabled  *bool   `json:"compaction_enabled,omitempty"`
	RetentionEnabled   *bool   `json:"retention_enabled,omitempty"`
}

// TopicLimits represents limits on topics.
type TopicLimits struct {
	MaxTopics *int `json:"max_topics,omitempty"`
}

// NewClusterTopicRequest provides a constructor to create or update a cluster topic.
func NewClusterTopicRequest(name string, replicationFactor int, retentionTime string) *clusterTopicRequest {
	return &clusterTopicRequest{Name: name, ReplicationFactor: replicationFactor, RetentionTime: retentionTime}
}

// clusterTopicRequest represents a request to create or modify a topic.
type clusterTopicRequest struct {
	// Name of the topic. Must not contain characters other than ASCII alphanumerics, '.', '_', and '-'
	Name string `json:"name"`

	// Number of partitions to give the topic.
	Partitions int `json:"partition_count,omitempty"`

	// Number of replicas the topic should be created across.
	ReplicationFactor int `json:"replication_factor"`

	// Length of time messages in the topic should be retained.
	// Minimum required is at least 24h or 86400000ms.
	// The client will convert the string value into a milliseconds integer value.
	//
	// Example: "10d". Supported suffixes:
	//  - `ms`, `millisecond`, `milliseconds`
	//  - `s`, `second`, `seconds`
	//  - `m`, `minute`, `minutes`
	//  - `h`, `hour`, `hours`
	//  - `d`, `day`, `days`
	RetentionTime string

	// Whether to use compaction for this topic.
	Compaction bool `json:"compaction"`

	// This field's retention value is used for the final request body.
	clusterTopicCreateRequestRetentionTime
}

type clusterTopicCreateRequestRetentionTime struct {
	RetentionTimeMS int64 `json:"retention_time_ms"`
}

type topicCreateRequest struct {
	Topic *clusterTopicRequest `json:"topic,omitempty"`
}

// ListClusterTopics returns a list of cluster topics.
func (k *Kafka) ListClusterTopics(clusterID string) (*Topics, *simpleresty.Response, error) {
	var result *Topics
	urlStr := k.http.RequestURL("/clusters/%s/topics", clusterID)

	// Execute the request
	response, getErr := k.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// GetClusterTopicByName finds a cluster topic by its name.
func (k *Kafka) GetClusterTopicByName(clusterID, topicName string) (*Topic, *simpleresty.Response, error) {
	topics, response, getErr := k.ListClusterTopics(clusterID)
	if getErr != nil {
		return nil, response, getErr
	}

	var topic *Topic
	for _, t := range topics.Topics {
		if t.GetName() == topicName {
			topic = t
		}
	}

	if topic == nil {
		return nil, nil, fmt.Errorf("no cluster topic named %s found on cluster %s", topicName, clusterID)
	}

	return topic, nil, nil
}

// CreateClusterTopic creates a cluster topic.
func (k *Kafka) CreateClusterTopic(clusterID string, opts *clusterTopicRequest) (*Response, *simpleresty.Response, error) {
	var result *Response
	urlStr := k.http.RequestURL("/clusters/%s/topics", clusterID)

	// Convert the retention value from string to integer
	rententionTimeMS, parseErr := convertDurationToMilliseconds(opts.RetentionTime)
	if parseErr != nil {
		return nil, nil, fmt.Errorf("unsupported retention time value")
	}
	opts.RetentionTimeMS = int64(rententionTimeMS)

	reqBody := &topicCreateRequest{
		Topic: opts,
	}

	// Execute the request
	response, createErr := k.http.Post(urlStr, &result, reqBody)

	return result, response, createErr
}

// UpdateClusterTopic updates an existing Kafka topic.
func (k *Kafka) UpdateClusterTopic(clusterID string, opts *clusterTopicRequest) (*Response, *simpleresty.Response, error) {
	var result *Response
	urlStr := k.http.RequestURL("/clusters/%s/topics/%s", clusterID, opts.Name)

	// Convert the retention value from string to integer
	rententionTimeMS, parseErr := convertDurationToMilliseconds(opts.RetentionTime)
	if parseErr != nil {
		return nil, nil, fmt.Errorf("unsupported retention time value")
	}
	opts.RetentionTimeMS = int64(rententionTimeMS)

	reqBody := &topicCreateRequest{
		Topic: opts,
	}

	// Execute the request
	response, updateErr := k.http.Put(urlStr, &result, reqBody)

	return result, response, updateErr
}

// DeleteClusterTopic deletes an existing topic.
func (k *Kafka) DeleteClusterTopic(clusterID, topicName string) (*Response, *simpleresty.Response, error) {
	var result *Response

	urlStr := k.http.RequestURL("/clusters/%s/topics/%s", clusterID, topicName)

	// Execute the request
	response, createErr := k.http.Delete(urlStr, &result, nil)

	return result, response, createErr
}
