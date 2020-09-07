package herokux

import (
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/davidji99/terraform-provider-herokux/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	DefaultMTLSProvisionTimeout         = int64(10)
	DefaultMTLSMTLSDeprovisionTimeout   = int64(10)
	DefaultMTLSIPRuleCreateTimeout      = int64(10)
	DefaultMTLSCertificateCreateTimeout = int64(10)
	DefaultMTLSCertificateDeleteTimeout = int64(10)
	DefaultKafkaCGCreateTimeout         = int64(10)
	DefaultKafkaCGDeleteTimeout         = int64(10)
)

type Config struct {
	API         *api.Client
	metricsURL  string
	postgresURL string
	token       string
	Headers     map[string]string

	// Custom Timeouts
	MTLSProvisionTimeout         int64
	MTLSDeprovisionTimeout       int64
	MTLSIPRuleCreateTimeout      int64
	MTLSCertificateCreateTimeout int64
	MTLSCertificateDeleteTimeout int64
	KafkaCGCreateTimeout         int64
	KafkaCGDeleteTimeout         int64
}

func NewConfig() *Config {
	c := &Config{
		MTLSProvisionTimeout:         DefaultMTLSProvisionTimeout,
		MTLSDeprovisionTimeout:       DefaultMTLSMTLSDeprovisionTimeout,
		MTLSIPRuleCreateTimeout:      DefaultMTLSIPRuleCreateTimeout,
		MTLSCertificateCreateTimeout: DefaultMTLSCertificateCreateTimeout,
		MTLSCertificateDeleteTimeout: DefaultMTLSCertificateDeleteTimeout,
		KafkaCGCreateTimeout:         DefaultKafkaCGCreateTimeout,
		KafkaCGDeleteTimeout:         DefaultKafkaCGDeleteTimeout,
	}
	return c
}

func (c *Config) initializeAPI() error {
	userAgent := fmt.Sprintf("terraform-provider-herokux/v%s", version.ProviderVersion)

	api, clientInitErr := api.New(config.APIToken(c.token), config.CustomHTTPHeaders(c.Headers),
		config.UserAgent(userAgent), config.MetricsBaseURL(c.metricsURL), config.PostgresBaseURL(c.postgresURL),
		config.BasicAuth("", c.token))
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

	if v, ok := d.GetOk("timeouts"); ok {
		vL := v.([]interface{})
		if len(vL) > 1 {
			return fmt.Errorf("provider configuration error: only 1 delays config is permitted")
		}
		for _, v := range vL {
			delaysConfig := v.(map[string]interface{})
			if v, ok := delaysConfig["mtls_provision_timeout"].(int); ok {
				c.MTLSProvisionTimeout = int64(v)
			}
			if v, ok := delaysConfig["mtls_deprovision_timeout"].(int); ok {
				c.MTLSDeprovisionTimeout = int64(v)
			}

			if v, ok := delaysConfig["mtls_iprule_create_timeout"].(int); ok {
				c.MTLSDeprovisionTimeout = int64(v)
			}

			if v, ok := delaysConfig["mtls_certificate_create_timeout"].(int); ok {
				c.MTLSCertificateCreateTimeout = int64(v)
			}

			if v, ok := delaysConfig["mtls_certificate_delete_timeout"].(int); ok {
				c.MTLSCertificateDeleteTimeout = int64(v)
			}

			if v, ok := delaysConfig["kafka_cg_create_timeout"].(int); ok {
				c.KafkaCGCreateTimeout = int64(v)
			}

			if v, ok := delaysConfig["kafka_cg_delete_timeout"].(int); ok {
				c.KafkaCGDeleteTimeout = int64(v)
			}
		}
	}

	return nil
}
