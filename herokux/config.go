package herokux

import (
	"fmt"
	"github.com/bgentry/go-netrc/netrc"
	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/davidji99/terraform-provider-herokux/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"
	"log"
	"net/http"
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
	DefaultDataConnectorCreateTimeout              = int64(10)
	DefaultDataConnectorDeleteTimeout              = int64(10)
	DefaultDataConnectorUpdateTimeout              = int64(10)
	DefaultPostgresCredentialCreateTimeout         = int64(10)
	DefaultPostgresCredentialDeleteTimeout         = int64(10)
	DefaultPostgresSettingsModifyDelay             = int64(2)
	DefaultPrivateSpaceCreateTimeout               = int64(20)
)

var (
	UserAgent = fmt.Sprintf("terraform-provider-herokux/v%s", version.ProviderVersion)
)

type Config struct {
	API         *api.Client
	PlatformAPI *heroku.Service
	platformURL string
	metricsURL  string
	postgresURL string
	dataURL     string
	redisURL    string
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
	DataConnectorCreateTimeout              int64
	DataConnectorDeleteTimeout              int64
	DataConnectorUpdateTimeout              int64
	PostgresCredentialCreateTimeout         int64
	PostgresCredentialDeleteTimeout         int64
	PrivateSpaceCreateTimeout               int64

	// Custom Delays
	PostgresSettingsModifyDelay int64
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
		DataConnectorCreateTimeout:              DefaultDataConnectorCreateTimeout,
		DataConnectorDeleteTimeout:              DefaultDataConnectorDeleteTimeout,
		DataConnectorUpdateTimeout:              DefaultDataConnectorUpdateTimeout,
		PostgresCredentialCreateTimeout:         DefaultPostgresCredentialCreateTimeout,
		PostgresCredentialDeleteTimeout:         DefaultPostgresCredentialDeleteTimeout,
		PostgresSettingsModifyDelay:             DefaultPostgresSettingsModifyDelay,
		PrivateSpaceCreateTimeout:               DefaultPrivateSpaceCreateTimeout,
	}
	return c
}

func (c *Config) initializeAPI() error {
	// Initialize the custom API client for non Heroku Platform APIs
	api, clientInitErr := api.New(config.APIToken(c.token), config.CustomHTTPHeaders(c.Headers),
		config.UserAgent(UserAgent), config.MetricsBaseURL(c.metricsURL), config.PostgresBaseURL(c.postgresURL),
		config.PlatformBaseURL(c.platformURL), config.RedisBaseURL(c.redisURL), config.BasicAuth("", c.token))
	if clientInitErr != nil {
		return clientInitErr
	}
	c.API = api

	// Initialize the Heroku Platform API client
	c.PlatformAPI = heroku.NewService(&http.Client{
		Transport: &heroku.Transport{
			Username:  "", // Email is not required
			Password:  c.token,
			UserAgent: UserAgent,
			Transport: heroku.RoundTripWithRetryBackoff{},
		},
	})
	c.PlatformAPI.URL = c.platformURL

	log.Printf("[INFO] Herokux Client configured")

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

	if v, ok := d.GetOk("data_api_url"); ok {
		vs := v.(string)
		c.dataURL = vs
	}

	if v, ok := d.GetOk("redis_api_url"); ok {
		vs := v.(string)
		c.redisURL = vs
	}

	if v, ok := d.GetOk("platform_api_url"); ok {
		vs := v.(string)
		c.platformURL = vs
	}

	if v, ok := d.GetOk("delays"); ok {
		vL := v.([]interface{})
		if len(vL) > 1 {
			return fmt.Errorf("provider configuration error: only 1 delays config is permitted")
		}
		for _, v := range vL {
			delaysConfig := v.(map[string]interface{})
			if v, ok := delaysConfig["postgres_settings_modify_delay"].(int); ok {
				c.PostgresSettingsModifyDelay = int64(v)
			}
		}
	}

	if v, ok := d.GetOk("timeouts"); ok {
		vL := v.([]interface{})
		if len(vL) > 1 {
			return fmt.Errorf("provider configuration error: only 1 timeout config is permitted")
		}
		for _, v := range vL {
			timeoutsConfig := v.(map[string]interface{})
			if v, ok := timeoutsConfig["mtls_provision_timeout"].(int); ok {
				c.MTLSProvisionTimeout = int64(v)
			}
			if v, ok := timeoutsConfig["mtls_deprovision_timeout"].(int); ok {
				c.MTLSDeprovisionTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["mtls_iprule_create_timeout"].(int); ok {
				c.MTLSDeprovisionTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["mtls_certificate_create_timeout"].(int); ok {
				c.MTLSCertificateCreateTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["mtls_certificate_delete_timeout"].(int); ok {
				c.MTLSCertificateDeleteTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["kafka_cg_create_timeout"].(int); ok {
				c.KafkaCGCreateTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["kafka_cg_delete_timeout"].(int); ok {
				c.KafkaCGDeleteTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["kafka_topic_create_timeout"].(int); ok {
				c.KafkaTopicCreateTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["kafka_topic_update_timeout"].(int); ok {
				c.KafkaTopicUpdateTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["privatelink_create_timeout"].(int); ok {
				c.PrivatelinkCreateTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["privatelink_delete_timeout"].(int); ok {
				c.PrivatelinkDeleteTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["privatelink_allowed_acccounts_add_timeout"].(int); ok {
				c.PrivatelinkAllowedAccountsAddTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["privatelink_allowed_acccounts_remove_timeout"].(int); ok {
				c.PrivatelinkAllowedAccountsRemoveTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["data_connector_create_timeout"].(int); ok {
				c.DataConnectorCreateTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["data_connector_delete_timeout"].(int); ok {
				c.DataConnectorDeleteTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["data_connector_update_timeout"].(int); ok {
				c.DataConnectorUpdateTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["postgres_credential_create_timeout"].(int); ok {
				c.PostgresCredentialCreateTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["postgres_credential_delete_timeout"].(int); ok {
				c.PostgresCredentialDeleteTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["shield_private_space_create_timeout"].(int); ok {
				c.PrivateSpaceCreateTimeout = int64(v)
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
