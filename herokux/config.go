package herokux

import (
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/davidji99/terraform-provider-herokux/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Config struct {
	API         *api.Client
	metricsURL  string
	postgresURL string
	token       string
	Headers     map[string]string
}

func NewConfig() *Config {
	config := &Config{}
	return config
}

func (c *Config) initializeAPI() error {
	userAgent := fmt.Sprintf("terraform-provider-herokux/v%s", version.ProviderVersion)

	api, clientInitErr := api.New(config.APIToken(c.token), config.CustomHTTPHeaders(c.Headers),
		config.UserAgent(userAgent), config.MetricsBaseURL(c.metricsURL), config.PostgresBaseURL(c.postgresURL))
	if clientInitErr != nil {
		return clientInitErr
	}

	c.API = api

	return nil
}

func (c *Config) applySchema(d *schema.ResourceData) (err error) {
	if v, ok := d.GetOk("headers"); ok {
		headersRaw := v.(map[string]interface{})
		h := make(map[string]string)

		for k, v := range headersRaw {
			h[k] = fmt.Sprintf("%v", v)
		}

		c.Headers = h
	}

	if v, ok := d.GetOk("metrics_api_url"); ok {
		vs := v.(string)
		c.metricsURL = vs
	}

	if v, ok := d.GetOk("postgres_api_url"); ok {
		vs := v.(string)
		c.postgresURL = vs
	}

	return nil
}
