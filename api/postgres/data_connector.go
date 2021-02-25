package postgres

import (
	"github.com/davidji99/simpleresty"
	"time"
)

// DataConnector represents a data connector in Heroku,
//
// Note: certain fields can be empty values depending on the request URL.
type DataConnector struct {
	ID              *string                     `json:"uuid,omitempty"`
	Name            *string                     `json:"name,omitempty"`
	Status          DataConnectorStatus         `json:"status,omitempty"`
	Lag             *string                     `json:"lag,omitempty"`
	CreatedAt       time.Time                   `json:"created_at,omitempty"`
	Settings        map[string]interface{}      `json:"settings"`
	Tables          []string                    `json:"tables"`
	ExcludedColumns []string                    `json:"excluded_columns"`
	KafkaApp        *DataConnectorKafkaApp      `json:"kafka_app,omitempty"`
	KafkaAddon      *DataConnectorKafkaAddon    `json:"kafka_addon,omitempty"`
	PostgresApp     *DataConnectorPostgresApp   `json:"postgres_app,omitempty"`
	PostgresAddon   *DataConnectorPostgresAddon `json:"postgres_addon,omitempty"`
	Topics          []*DataConnectorTopic       `json:"topics,omitempty"`
}

// DataConnectorKafkaApp represents a data connector's Kafka app.
type DataConnectorKafkaApp struct {
	Name *string `json:"name,omitempty"`
}

// DataConnectorKafkaAddon represents a data connector's Kafka addon.
type DataConnectorKafkaAddon struct {
	Name *string `json:"name,omitempty"`
	ID   *string `json:"uuid,omitempty"`
}

// DataConnectorPostgresApp represents a data connector's Postgres app.
type DataConnectorPostgresApp struct {
	Name *string `json:"name,omitempty"`
}

// DataConnectorPostgresAddon represents a data connector's Postgres addon.
type DataConnectorPostgresAddon struct {
	Name *string `json:"name,omitempty"`
	ID   *string `json:"uuid,omitempty"`
}

// DataConnectorTopic represents a data connector's kafka topics.
type DataConnectorTopic struct {
	TableName *string `json:"table_name,omitempty"`
	TopicName *string `json:"topic_name,omitempty"`
}

// NewDataConnectorRequest represents a new request to create a data connector.
func NewDataConnectorRequest(postgresAddonID string, tables []string) *dataConnectorRequest {
	return &dataConnectorRequest{
		PostgresAddonID: postgresAddonID,
		Tables:          tables,
		ExcludedColumns: make([]string, 0),
		PlatformVersion: "",
		Name:            "",
	}
}

// dataConnectorRequest represents a request to create a data connector.
type dataConnectorRequest struct {
	// PostgresAddonID
	PostgresAddonID string `json:"postgres_addon_uuid"`

	// Tables
	Tables []string `json:"tables"`

	// ExcludedColumns excludes columns from a table for streaming. Must be at least an empty array.
	ExcludedColumns []string `json:"excluded_columns"`

	// PlatformVersion. Must be at least an empty string.
	PlatformVersion string `json:"platform_version"`

	// Name of the data connector. Must be at least an empty string.
	Name string `json:"name"`
}

// DataConnectSettings represents properties associated with your connector.
type DataConnectSettings struct {
	Settings map[string]interface{} `json:"settings"`
}

// ListDataConnectors retrieves all data connectors for an app.
func (p *Postgres) ListDataConnectors(nameOrID string) ([]*DataConnector, *simpleresty.Response, error) {
	var result []*DataConnector
	urlStr := p.http.RequestURL("/data/cdc/v0/apps/%s", nameOrID)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// GetDataConnector retrieves a single data connector.
func (p *Postgres) GetDataConnector(id string) (*DataConnector, *simpleresty.Response, error) {
	var result *DataConnector
	urlStr := p.http.RequestURL("/data/cdc/v0/connectors/%s", id)

	// Execute the request
	response, getErr := p.http.Get(urlStr, &result, nil)

	return result, response, getErr
}

// CreateDataConnector creates a data connector.
func (p *Postgres) CreateDataConnector(kafkaID string, opts *dataConnectorRequest) (*DataConnector, *simpleresty.Response, error) {
	var result *DataConnector
	urlStr := p.http.RequestURL("/data/cdc/v0/kafka_tenants/%s", kafkaID)

	// Execute the request
	response, createErr := p.http.Post(urlStr, &result, opts)

	return result, response, createErr
}

// DeleteDataConnector deletes a data connector.
//
// The deleted data connector is returned. It is likely that the connector is fully deleted
// when the GET/{id} returns a 404.
//
// Heroku does not delete your Kafka topics automatically, because they could still contain messages
// which you haven't consumed. Please delete the topics manually.
func (p *Postgres) DeleteDataConnector(id string) (*DataConnector, *simpleresty.Response, error) {
	var result *DataConnector
	urlStr := p.http.RequestURL("/data/cdc/v0/connectors/%s", id)

	// Execute the request
	response, deleteErr := p.http.Delete(urlStr, &result, nil)

	return result, response, deleteErr
}

// PauseDataConnector pauses change event creation on a Data Connector.
//
// Returns 202 with empty body if accepted and takes a bit of time before the pause is applied to the data connector.
func (p *Postgres) PauseDataConnector(id string) (*simpleresty.Response, error) {
	urlStr := p.http.RequestURL("/data/cdc/v0/connectors/%s/pause", id)

	opts := struct {
		Raw bool `json:"raw"`
	}{
		Raw: true,
	}

	// Execute the request
	response, pauseErr := p.http.Put(urlStr, nil, &opts)

	return response, pauseErr
}

// ResumeDataConnector resumes change event creation on a Data Connector.
//
// Returns 202 with empty body if accepted and takes a bit of time before the resume is applied to the data connector.
func (p *Postgres) ResumeDataConnector(id string) (*simpleresty.Response, error) {
	urlStr := p.http.RequestURL("/data/cdc/v0/connectors/%s/resume", id)

	opts := struct {
		Raw bool `json:"raw"`
	}{
		Raw: true,
	}

	// Execute the request
	response, resume := p.http.Put(urlStr, nil, &opts)

	return response, resume
}

// UpdateDataConnectorSettings updates the settings for a Data Connector.
//
// Reference: https://devcenter.heroku.com/articles/heroku-data-connectors#update-configuration
func (p *Postgres) UpdateDataConnectorSettings(id string, opts *DataConnectSettings) (*DataConnector, *simpleresty.Response, error) {
	var result DataConnector
	urlStr := p.http.RequestURL("/data/cdc/v0/connectors/%s", id)

	// Execute the request
	response, resume := p.http.Patch(urlStr, &result, opts)

	return &result, response, resume
}
