package herokuplus

import (
	"fmt"
	"github.com/davidji99/terraform-provider-herokuplus/api"
	"github.com/davidji99/terraform-provider-herokuplus/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Config struct {
	API     *api.Client
	url     string
	token   string
	Headers map[string]string
}

func NewConfig() *Config {
	config := &Config{}
	return config
}

func (c *Config) initializeAPI() error {
	userAgent := fmt.Sprintf("terraform-provider-herokuplus/v%s", version.ProviderVersion)

	api, clientInitErr := api.New(api.APIToken(c.token), api.CustomHTTPHeaders(c.Headers),
		api.UserAgent(userAgent), api.BaseURL(c.url))
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

	if v, ok := d.GetOk("url"); ok {
		vs := v.(string)
		c.url = vs
	}

	return nil
}
