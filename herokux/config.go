package herokux

import (
	"fmt"
	"github.com/bgentry/go-netrc/netrc"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/davidji99/terraform-provider-herokux/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	heroku "github.com/heroku/heroku-go/v5"
	"github.com/mitchellh/go-homedir"
	"net/url"
	"os"
	"runtime"
)

const (
	DefaultMTLSProvisionTimeout                    = int64(10)
	DefaultMTLSMTLSDeprovisionTimeout              = int64(10)
	DefaultMTLSIPRuleCreateTimeout                 = int64(10)
	DefaultMTLSCertificateCreateTimeout            = int64(10)
	DefaultMTLSCertificateDeleteTimeout            = int64(10)
	DefaultKafkaCGCreateTimeout                    = int64(10)
	DefaultKafkaCGDeleteTimeout                    = int64(10)
	DefaultKafkaTopicCreateTimeout                 = int64(10)
	DefaultKafkaTopicUpdateTimeout                 = int64(10)
	DefaultPrivatelinkCreateTimeoutt               = int64(15)
	DefaultPrivatelinkDeleteTimeout                = int64(15)
	DefaultPrivatelinkAllowedAccountsAddTimeout    = int64(10)
	DefaultPrivatelinkAllowedAccountsRemoveTimeout = int64(10)
)

type Config struct {
	API         *api.Client
	metricsURL  string
	postgresURL string
	token       string
	Headers     map[string]string

	// Custom Timeouts
	MTLSProvisionTimeout                    int64
	MTLSDeprovisionTimeout                  int64
	MTLSIPRuleCreateTimeout                 int64
	MTLSCertificateCreateTimeout            int64
	MTLSCertificateDeleteTimeout            int64
	KafkaCGCreateTimeout                    int64
	KafkaCGDeleteTimeout                    int64
	KafkaTopicCreateTimeout                 int64
	KafkaTopicUpdateTimeout                 int64
	PrivatelinkCreateTimeout                int64
	PrivatelinkDeleteTimeout                int64
	PrivatelinkAllowedAccountsAddTimeout    int64
	PrivatelinkAllowedAccountsRemoveTimeout int64
}

func NewConfig() *Config {
	c := &Config{
		MTLSProvisionTimeout:                    DefaultMTLSProvisionTimeout,
		MTLSDeprovisionTimeout:                  DefaultMTLSMTLSDeprovisionTimeout,
		MTLSIPRuleCreateTimeout:                 DefaultMTLSIPRuleCreateTimeout,
		MTLSCertificateCreateTimeout:            DefaultMTLSCertificateCreateTimeout,
		MTLSCertificateDeleteTimeout:            DefaultMTLSCertificateDeleteTimeout,
		KafkaCGCreateTimeout:                    DefaultKafkaCGCreateTimeout,
		KafkaCGDeleteTimeout:                    DefaultKafkaCGDeleteTimeout,
		KafkaTopicCreateTimeout:                 DefaultKafkaTopicCreateTimeout,
		KafkaTopicUpdateTimeout:                 DefaultKafkaTopicUpdateTimeout,
		PrivatelinkCreateTimeout:                DefaultPrivatelinkCreateTimeoutt,
		PrivatelinkDeleteTimeout:                DefaultPrivatelinkDeleteTimeout,
		PrivatelinkAllowedAccountsAddTimeout:    DefaultPrivatelinkAllowedAccountsAddTimeout,
		PrivatelinkAllowedAccountsRemoveTimeout: DefaultPrivatelinkAllowedAccountsRemoveTimeout,
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

			if v, ok := delaysConfig["kafka_topic_create_timeout"].(int); ok {
				c.KafkaTopicCreateTimeout = int64(v)
			}

			if v, ok := delaysConfig["kafka_topic_update_timeout"].(int); ok {
				c.KafkaTopicUpdateTimeout = int64(v)
			}

			if v, ok := delaysConfig["privatelink_create_timeout"].(int); ok {
				c.PrivatelinkCreateTimeout = int64(v)
			}

			if v, ok := delaysConfig["privatelink_delete_timeout"].(int); ok {
				c.PrivatelinkDeleteTimeout = int64(v)
			}

			if v, ok := delaysConfig["privatelink_allowed_acccounts_add_timeout"].(int); ok {
				c.PrivatelinkAllowedAccountsAddTimeout = int64(v)
			}

			if v, ok := delaysConfig["privatelink_allowed_acccounts_remove_timeout"].(int); ok {
				c.PrivatelinkAllowedAccountsRemoveTimeout = int64(v)
			}
		}
	}

	return nil
}

func (c *Config) applyNetrcFile() error {
	// Get the netrc file path. If path not shown, then fall back to default netrc path value
	path := os.Getenv("NETRC_PATH")

	if path == "" {
		filename := ".netrc"
		if runtime.GOOS == "windows" {
			filename = "_netrc"
		}

		var err error
		path, err = homedir.Expand("~/" + filename)
		if err != nil {
			return err
		}
	}

	// If the file is not a file, then do nothing
	if fi, err := os.Stat(path); err != nil {
		// File doesn't exist, do nothing
		if os.IsNotExist(err) {
			return nil
		}

		// Some other error!
		return err
	} else if fi.IsDir() {
		// File is directory, ignore
		return nil
	}

	// Load up the netrc file
	net, err := netrc.ParseFile(path)
	if err != nil {
		return fmt.Errorf("error parsing netrc file at %q: %s", path, err)
	}

	// Reference the default Heroku Platform API url from heroku-go as that's the host URL used in ~/.netrc.
	// Doing this is okay because although this provider uses different base endpoints,
	//the authentication among all of the endpoints.
	u, err := url.Parse(heroku.DefaultURL)
	if err != nil {
		return err
	}

	machine := net.FindMachine(u.Host)
	if machine == nil {
		// Machine not found, no problem
		return nil
	}

	c.token = machine.Password

	return nil
}
