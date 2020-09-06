package kafka

import (
	"github.com/davidji99/simpleresty"
	"time"
)

// Cluster represents a Kafka cluster.
type Cluster struct {
	AddonID                  *string              `json:"addon_id,omitempty"`
	Name                     *string              `json:"name,omitempty"`
	AddonAttachmentConfigVar *string              `json:"attachment_name,omitempty"`
	CreatedAt                *time.Time           `json:"created_at,omitempty"`
	State                    *ClusterState        `json:"state,omitempty"`
	Limits                   *ClusterLimits       `json:"limits,omitempty"`
	MetaasSource             *string              `json:"metaas_source,omitempty"`
	MessagesInPerSec         *int64               `json:"messages_in_per_sec,omitempty"`
	BytesInPerSec            *int64               `json:"bytes_in_per_sec,omitempty"`
	BytesOutPerSec           *int64               `json:"bytes_out_per_sec,omitempty"`
	Versions                 []string             `json:"version,omitempty"`
	PartitionReplicaCount    *int                 `json:"partition_replica_count,omitempty"`
	DataSize                 *int64               `json:"data_size,omitempty"`
	Topics                   []string             `json:"topics,omitempty"`
	AdminTopicNames          []string             `json:"admin_topic_names,omitempty"`
	SharedCluster            *bool                `json:"shared_cluster,omitempty"`
	TopicPrefix              *string              `json:"topic_prefix,omitempty"`
	Formation                *ClusterFormation    `json:"formation,omitempty"`
	Defaults                 *ClusterDefaults     `json:"defaults,omitempty"`
	Robot                    *ClusterRobot        `json:"robot,omitempty"`
	Capabilities             *ClusterCapabilities `json:"capabilities,omitempty"`
}

// ClusterDefaults represents information about a cluster's defaults.
type ClusterDefaults struct {
	PartitionCount  *int   `json:"partition_count,omitempty"`
	RetentionTimeMS *int64 `json:"retention_time_ms,omitempty"`
}

// ClusterDefaults represents information about a cluster's robot.
type ClusterRobot struct {
	IsRobot  *bool `json:"is_robot,omitempty"`
	RobotTTL *int  `json:"robot_ttl,omitempty"`
}

// ClusterCapabilities represents information about a cluster's capabilities.
type ClusterCapabilities struct {
	SupportsMixedCleanupPolicy *bool `json:"supports_mixed_cleanup_policy,omitempty"`
}

// ClusterFormation represents information about a cluster's formation.
type ClusterFormation struct {
	ID           *string  `json:"id,omitempty"`
	KafkaIDs     []string `json:"kafkas,omitempty"`
	ZookeeperIDs []string `json:"zookeepers,omitempty"`
}

// ClusterState represents information about a cluster's state.
type ClusterState struct {
	Message         *string  `json:"message,omitempty"`
	Waiting         *bool    `json:"waiting,omitempty"`
	Healthy         *bool    `json:"healthy,omitempty"`
	Status          *string  `json:"status,omitempty"`
	DegradedTopics  []string `json:"degraded_topics,omitempty"`
	DegradedBrokers []string `json:"degraded_brokers,omitempty"`
}

// ClusterLimits represents information about a cluster's limits.
type ClusterLimits struct {
	MinimumReplication            *int                   `json:"minimum_replication,omitempty"`
	MaximumReplication            *int                   `json:"maximum_replication,omitempty"`
	MinimumReplicationMS          *int64                 `json:"minimum_retention_ms,omitempty"`
	MaximumReplicationMS          *int64                 `json:"maximum_retention_ms,omitempty"`
	MaxPartitionReplicaCount      *int                   `json:"max_partition_replica_count,omitempty"`
	MaxNumberOfTotalPartitions    *int                   `json:"max_number_of_total_partitions,omitempty"`
	MaxNumberOfPartitionsPerTopic *int                   `json:"max_number_of_partitions_per_topic,omitempty"`
	MaxNumberOfTopics             *int                   `json:"max_number_of_topics,omitempty"`
	DataSize                      *ClusterLimitsDataSize `json:"data_size,omitempty"`
	ProduceQuotaBytesPerSecond    *int64                 `json:"produce_quota_bytes_per_second,omitempty"`
	ConsumeQuotaBytesPerSecond    *int64                 `json:"consume_quota_bytes_per_second,omitempty"`
	MaxTopics                     *int                   `json:"max_topics,omitempty"`
}

// ClusterLimitsDataSize represents information about a cluster limit's data size.
type ClusterLimitsDataSize struct {
	SubCriticalPercentage   *int   `json:"subcritical_percentage,omitempty"`
	CriticalPercentage      *int   `json:"critical_percentage,omitempty"`
	SuperCriticalPercentage *int   `json:"supercritical_percentage,omitempty"`
	LimitByte               *int64 `json:"limit_bytes,omitempty"`
}

// Get retrieves information about a Kafka cluster.
func (k *Kafka) Get(clusterID string) (*Cluster, *simpleresty.Response, error) {
	var result *Cluster
	urlStr := k.http.RequestURL("/clusters/%s", clusterID)

	// Execute the request
	response, getErr := k.http.Get(urlStr, &result, nil)

	return result, response, getErr
}
