package kafka

// ConsumerGroupStatus represent the status of a consumer group.
type ConsumerGroupStatus string

// ConsumerGroupStatuses represent all statuses pertaining to the lifecycle of a consumer group.
var ConsumerGroupStatuses = struct {
	CREATED ConsumerGroupStatus
	PENDING ConsumerGroupStatus
	DELETED ConsumerGroupStatus
	UNKNOWN ConsumerGroupStatus
}{
	CREATED: "created",
	PENDING: "pending",
	DELETED: "deleted",
	UNKNOWN: "Unknown",
}

// ToString is a helper method to return the string of a ConsumerGroupStatus.
func (s ConsumerGroupStatus) ToString() string {
	return string(s)
}

// TopicStatus represent the status of a topic
type TopicStatus string

// TopicStatuses represent all statuses pertaining to the lifecycle of a topic
var TopicStatuses = struct {
	PENDING  TopicStatus
	CREATED  TopicStatus
	READY    TopicStatus
	UPDATING TopicStatus
	UPDATED  TopicStatus
	DELETED  TopicStatus
	UNKNOWN  TopicStatus
}{
	PENDING:  "pending",
	CREATED:  "created",
	READY:    "ready",
	UPDATING: "updating",
	UPDATED:  "updated",
	DELETED:  "deleted",
	UNKNOWN:  "Unknown",
}

// ToString is a helper method to return the string of a TopicStatus.
func (s TopicStatus) ToString() string {
	return string(s)
}
