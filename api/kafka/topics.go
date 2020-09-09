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
	ReplicationFactor  *int    `json:"replication_factor,omitempty"`
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

// TopicRequest represents a request to create or modify a topic.
type TopicRequest struct {
	// Name of the topic. Must not contain characters other than ASCII alphanumerics, '.', '_', and '-'
	Name string `json:"name,omitempty"`

	// Number of partitions to give the topic.
	Partitions int `json:"partition_count,omitempty"`

	// Number of replicas the topic should be created across.
	ReplicationFactor int `json:"replication_factor,omitempty"`

	// Length of time messages in the topic should be retained.
	// Minimum required is at least 24h or 86400000ms.
	RetentionTimeMS *int `json:"retention_time_ms"`

	// Whether to use compaction for this topic.
	Compaction bool `json:"compaction"`
}

type topicRequestBody struct {
	Topic *TopicRequest `json:"topic,omitempty"`
}

// ListTopics returns a list of cluster topics.
func (k *Kafka) ListTopics(clusterID string) (*Topics, *simpleresty.Response, error) {
	var result *Topics
	urlStr := k.http.RequestURL("/clusters/%s/topics", clusterID)

	// Execute the request
	response, getErr := k.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// GetTopicByName finds a cluster topic by its name.
func (k *Kafka) GetTopicByName(clusterID, topicName string) (*Topic, *simpleresty.Response, error) {
	topics, response, getErr := k.ListTopics(clusterID)
	if getErr != nil {
		return nil, response, getErr
	}

	var topic *Topic
	for _, t := range topics.Topics {
		if t.GetName() == topicName {
			topic = t
		}
	}

	r := &simpleresty.Response{}
	if topic == nil {
		r.StatusCode = 404
		return nil, r, fmt.Errorf("no cluster topic named %s found on cluster %s", topicName, clusterID)
	}

	return topic, r, nil
}

// CreateTopic creates a cluster topic.
func (k *Kafka) CreateTopic(clusterID string, opts *TopicRequest) (*Response, *simpleresty.Response, error) {
	var result *Response
	urlStr := k.http.RequestURL("/clusters/%s/topics", clusterID)
	reqBody := &topicRequestBody{Topic: opts}

	// Execute the request
	response, createErr := k.http.Post(urlStr, &result, reqBody)

	return result, response, createErr
}

// UpdateTopic updates an existing Kafka topic.
func (k *Kafka) UpdateTopic(clusterID string, opts *TopicRequest) (*Response, *simpleresty.Response, error) {
	var result *Response
	urlStr := k.http.RequestURL("/clusters/%s/topics/%s", clusterID, opts.Name)
	reqBody := &topicRequestBody{Topic: opts}

	// Execute the request
	response, updateErr := k.http.Put(urlStr, &result, reqBody)

	return result, response, updateErr
}

// DeleteTopic deletes an existing topic.
func (k *Kafka) DeleteTopic(clusterID, topicName string) (*Response, *simpleresty.Response, error) {
	var result *Response

	urlStr := k.http.RequestURL("/clusters/%s/topics/%s", clusterID, topicName)

	// Execute the request
	response, createErr := k.http.Delete(urlStr, &result, nil)

	return result, response, createErr
}
