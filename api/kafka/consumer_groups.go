package kafka

import (
	"fmt"
	"github.com/davidji99/simpleresty"
)

// ConsumerGroups represents a list of consumer groups.
type ConsumerGroups struct {
	AddonAttachmentConfigVar *string          `json:"attachment_name,omitempty"`
	ConsumerGroups           []*ConsumerGroup `json:"consumer_groups,omitempty"`
}

// ConsumerGroup represents a consumer group.
type ConsumerGroup struct {
	Name *string
}

// NewConsumerGroupRequest defines a constructor to create or destroy a consumer group.
func NewConsumerGroupRequest() *consumerGroupRequest {
	return &consumerGroupRequest{}
}

type consumerGroupRequest struct {
	Name string `json:"name"`
}

type consumerGroupBody struct {
	ConsumerGroup *consumerGroupRequest `json:"consumer_group,omitempty"`
}

// ListConsumerGroups returns a list of all consumer groups.
func (k *Kafka) ListConsumerGroups(clusterID string) (*ConsumerGroups, *simpleresty.Response, error) {
	var result *ConsumerGroups
	urlStr := k.http.RequestURL("/data/kafka/v0/clusters/%s/consumer_groups", clusterID)

	// Execute the request
	response, getErr := k.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

func (k *Kafka) GetConsumerGroupByName(clusterID, groupName string) (*ConsumerGroup, *simpleresty.Response, error) {
	groups, _, listErr := k.ListConsumerGroups(clusterID)
	if listErr != nil {
		return nil, nil, listErr
	}

	for _, g := range groups.ConsumerGroups {
		if g.GetName() == groupName {
			return g, nil, nil
		}
	}

	return nil, nil, fmt.Errorf("%s not found", groupName)
}

// CreateConsumerGroup creates a single consumer group.
//
// Requests to create duplicate groups result in a no-operation.
// The group is not ready for use until it appears in the LIST response.
func (k *Kafka) CreateConsumerGroup(clusterID string, opts *consumerGroupRequest) (*Response, *simpleresty.Response, error) {
	var result *Response
	urlStr := k.http.RequestURL("/data/kafka/v0/clusters/%s/consumer_groups", clusterID)

	reqBody := &consumerGroupBody{
		ConsumerGroup: opts,
	}

	// Execute the request
	response, createErr := k.http.Post(urlStr, &result, reqBody)

	return result, response, createErr
}

// DeleteConsumerGroup deletes an existing consumer group.
func (k *Kafka) DeleteConsumerGroup(clusterID string, opts *consumerGroupRequest) (*Response, *simpleresty.Response, error) {
	var result *Response
	urlStr := k.http.RequestURL("/data/kafka/v0/clusters/%s/consumer_groups", clusterID)

	reqBody := &consumerGroupBody{
		ConsumerGroup: opts,
	}

	// Execute the request
	response, deleteErr := k.http.Delete(urlStr, &result, reqBody)

	return result, response, deleteErr
}

// WasConsumerGroupCreated provides a simple method to determine if a consumer group was created successfully.
//
// This check is done by determining whether the consumer group is present when listing all consumer groups.
func (k *Kafka) WasConsumerGroupCreated(clusterID, consumerGroupName string) (bool, *simpleresty.Response, error) {
	listResp, response, listErr := k.ListConsumerGroups(clusterID)
	if listErr != nil {
		return false, response, listErr
	}

	for _, cg := range listResp.ConsumerGroups {
		if cg.GetName() == consumerGroupName {
			return true, nil, nil
		}
	}

	return false, nil, nil
}

// WasConsumerGroupDeleted provides a simple method to determine if a consumer group was deleted successfully.
//
// This check is done by determining whether the consumer group is not present when listing all consumer groups.
func (k *Kafka) WasConsumerGroupDeleted(clusterID, consumerGroupName string) (bool, *simpleresty.Response, error) {
	isCreated, response, err := k.WasConsumerGroupCreated(clusterID, consumerGroupName)
	if err != nil {
		return false, response, err
	}

	return !isCreated, nil, nil
}
