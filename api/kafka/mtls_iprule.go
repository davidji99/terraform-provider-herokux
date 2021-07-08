package kafka

import (
	"github.com/davidji99/simpleresty"
	"github.com/davidji99/terraform-provider-herokux/api/general"
)

// ListMTLSIPRules returns all IP rules.
func (k *Kafka) ListMTLSIPRules(kafkaID string) ([]*general.MtlsIPRule, *simpleresty.Response, error) {
	var result []*general.MtlsIPRule
	urlStr := k.http.RequestURL("/data/kafka/v0/clusters/%s/ip-rules", kafkaID)

	// Execute the request
	response, getErr := k.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// GetMTLSIPRule returns a single IP rule.
func (k *Kafka) GetMTLSIPRule(kafkaID, ruleID string) ([]*general.MtlsIPRule, *simpleresty.Response, error) {
	var result []*general.MtlsIPRule
	urlStr := k.http.RequestURL("/data/kafka/v0/clusters/%s/ip-rules/%s", kafkaID, ruleID)

	// Execute the request
	response, getErr := k.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// CreateMTLSIPRule creates a single IP rule.
func (k *Kafka) CreateMTLSIPRule(kafkaID string, opts *general.MTLSIPRuleRequest) (*general.MtlsIPRule, *simpleresty.Response, error) {
	var result general.MtlsIPRule
	urlStr := k.http.RequestURL("/data/kafka/v0/clusters/%s/ip-rules", kafkaID)

	// Execute the request
	response, getErr := k.http.Post(urlStr, &result, opts)

	return &result, response, getErr
}

// DeleteMTLSIPRule deletes a single IP rule.
func (k *Kafka) DeleteMTLSIPRule(kafkaID, ruleID string) (*simpleresty.Response, error) {
	urlStr := k.http.RequestURL("/data/kafka/v0/clusters/%s/ip-rules/%s", kafkaID, ruleID)

	// Execute the request
	response, getErr := k.http.Delete(urlStr, nil, nil)

	return response, getErr
}
