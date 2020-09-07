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
