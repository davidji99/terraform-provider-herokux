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
	DefaultMTLSProvisionVerifyTimeout                    = int64(10)
	DefaultMTLSMTLSDeprovisionVerifyTimeout              = int64(10)
	DefaultMTLSIPRuleCreateVerifyTimeout                 = int64(20)
	DefaultMTLSCertificateCreateVerifyTimeout            = int64(10)
	DefaultMTLSCertificateDeleteVerifyTimeout            = int64(10)
	DefaultKafkaCGCreateVerifyTimeout                    = int64(10)
	DefaultKafkaCGDeleteVerifyTimeout                    = int64(10)
	DefaultKafkaTopicCreateVerifyTimeout                 = int64(10)
	DefaultKafkaTopicUpdateVerifyTimeout                 = int64(10)
	DefaultPrivatelinkCreateVerifyTimeout                = int64(15)
	DefaultPrivatelinkDeleteVerifyTimeout                = int64(15)
	DefaultPrivatelinkAllowedAccountsAddVerifyTimeout    = int64(10)
	DefaultPrivatelinkAllowedAccountsRemoveVerifyTimeout = int64(10)
	DefaultDataConnectorCreateVerifyTimeout              = int64(20)
	DefaultDataConnectorSettingsUpdateVerifyTimeout      = int64(10)
	DefaultDataConnectorDeleteVerifyTimeout              = int64(10)
	DefaultDataConnectorStatusUpdateVerifyTimeout        = int64(10)
	DefaultPostgresCredentialPreCreateVerifyTimeout      = int64(45)
	DefaultPostgresCredentialCreateVerifyTimeout         = int64(10)
	DefaultPostgresCredentialDeleteVerifyTimeout         = int64(10)
	DefaultPrivateSpaceCreateVerifyTimeout               = int64(20)
	DefaultAppContainerReleaseVerifyTimeout              = int64(20)

	DefaultPostgresSettingsModifyDelay = int64(2)
	DefaultConnectMappingModifyDelay   = int64(15)
)

var (
	UserAgent = fmt.Sprintf("terraform-provider-herokux/v%s", version.ProviderVersion)
)

type Config struct {
	API         *api.Client
	PlatformAPI *heroku.Service
	token       string
	Headers     map[string]string

	// Custom API URLs
	platformURL       string
	metricsURL        string
	postgresURL       string
	dataURL           string
	redisURL          string
	connectCentralURL string
	registryURL       string
	kolkrabbiURL      string
	schedulerURL      string

	// Custom Timeouts
	MTLSProvisionVerifyTimeout                    int64
	MTLSDeprovisionVerifyTimeout                  int64
	MTLSIPRuleCreateVerifyTimeout                 int64
	MTLSCertificateCreateVerifyTimeout            int64
	MTLSCertificateDeleteVerifyTimeout            int64
	KafkaCGCreateVerifyTimeout                    int64
	KafkaCGDeleteVerifyTimeout                    int64
	KafkaTopicCreateVerifyTimeout                 int64
	KafkaTopicUpdateVerifyTimeout                 int64
	PrivatelinkCreateVerifyTimeout                int64
	PrivatelinkDeleteVerifyTimeout                int64
	PrivatelinkAllowedAccountsAddVerifyTimeout    int64
	PrivatelinkAllowedAccountsRemoveVerifyTimeout int64
	DataConnectorCreateVerifyTimeout              int64
	DataConnectorSettingsUpdateVerifyTimeout      int64
	DataConnectorDeleteVerifyTimeout              int64
	DataConnectorStatusUpdateVerifyTimeout        int64
	PostgresCredentialCreateVerifyTimeout         int64
	PostgresCredentialPreCreateVerifyTimeout      int64
	PostgresCredentialDeleteVerifyTimeout         int64
	PrivateSpaceCreateVerifyTimeout               int64
	AppContainerReleaseVerifyTimeout              int64

	// Custom Delays
	PostgresSettingsModifyDelay int64
	ConnectMappingModifyDelay   int64
}

func NewConfig() *Config {
	c := &Config{
		MTLSProvisionVerifyTimeout:                    DefaultMTLSProvisionVerifyTimeout,
		MTLSDeprovisionVerifyTimeout:                  DefaultMTLSMTLSDeprovisionVerifyTimeout,
		MTLSIPRuleCreateVerifyTimeout:                 DefaultMTLSIPRuleCreateVerifyTimeout,
		MTLSCertificateCreateVerifyTimeout:            DefaultMTLSCertificateCreateVerifyTimeout,
		MTLSCertificateDeleteVerifyTimeout:            DefaultMTLSCertificateDeleteVerifyTimeout,
		KafkaCGCreateVerifyTimeout:                    DefaultKafkaCGCreateVerifyTimeout,
		KafkaCGDeleteVerifyTimeout:                    DefaultKafkaCGDeleteVerifyTimeout,
		KafkaTopicCreateVerifyTimeout:                 DefaultKafkaTopicCreateVerifyTimeout,
		KafkaTopicUpdateVerifyTimeout:                 DefaultKafkaTopicUpdateVerifyTimeout,
		PrivatelinkCreateVerifyTimeout:                DefaultPrivatelinkCreateVerifyTimeout,
		PrivatelinkDeleteVerifyTimeout:                DefaultPrivatelinkDeleteVerifyTimeout,
		PrivatelinkAllowedAccountsAddVerifyTimeout:    DefaultPrivatelinkAllowedAccountsAddVerifyTimeout,
		PrivatelinkAllowedAccountsRemoveVerifyTimeout: DefaultPrivatelinkAllowedAccountsRemoveVerifyTimeout,
		DataConnectorCreateVerifyTimeout:              DefaultDataConnectorCreateVerifyTimeout,
		DataConnectorSettingsUpdateVerifyTimeout:      DefaultDataConnectorSettingsUpdateVerifyTimeout,
		DataConnectorDeleteVerifyTimeout:              DefaultDataConnectorDeleteVerifyTimeout,
		DataConnectorStatusUpdateVerifyTimeout:        DefaultDataConnectorStatusUpdateVerifyTimeout,
		PostgresCredentialPreCreateVerifyTimeout:      DefaultPostgresCredentialPreCreateVerifyTimeout,
		PostgresCredentialCreateVerifyTimeout:         DefaultPostgresCredentialCreateVerifyTimeout,
		PostgresCredentialDeleteVerifyTimeout:         DefaultPostgresCredentialDeleteVerifyTimeout,
		PrivateSpaceCreateVerifyTimeout:               DefaultPrivateSpaceCreateVerifyTimeout,
		AppContainerReleaseVerifyTimeout:              DefaultAppContainerReleaseVerifyTimeout,

		PostgresSettingsModifyDelay: DefaultPostgresSettingsModifyDelay,
		ConnectMappingModifyDelay:   DefaultConnectMappingModifyDelay,
	}
	return c
}

func (c *Config) initializeAPI() error {
	// Initialize the custom API client for non Heroku Platform APIs
	api, clientInitErr := api.New(config.APIToken(c.token), config.BasicAuth("", c.token),
		config.CustomHTTPHeaders(c.Headers),
		config.UserAgent(UserAgent),
		config.MetricsBaseURL(c.metricsURL),
		config.PostgresBaseURL(c.postgresURL),
		config.PlatformBaseURL(c.platformURL),
		config.RedisBaseURL(c.redisURL),
		config.ConnectCentralBaseURL(c.connectCentralURL),
		config.RegistryBaseURL(c.registryURL),
		config.KolkrabbiBaseURL(c.kolkrabbiURL),
		config.SchedulerBaseURL(c.schedulerURL),
	)
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

	if v, ok := d.GetOk("connect_central_api_url"); ok {
		vs := v.(string)
		c.connectCentralURL = vs
	}

	if v, ok := d.GetOk("registry_api_url"); ok {
		vs := v.(string)
		c.registryURL = vs
	}

	if v, ok := d.GetOk("kolkrabbi_api_url"); ok {
		vs := v.(string)
		c.kolkrabbiURL = vs
	}

	if v, ok := d.GetOk("scheduler_api_url"); ok {
		vs := v.(string)
		c.schedulerURL = vs
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

			if v, ok := delaysConfig["connect_mapping_modify_delay"].(int); ok {
				c.ConnectMappingModifyDelay = int64(v)
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
			if v, ok := timeoutsConfig["mtls_provision_verify_timeout"].(int); ok {
				c.MTLSProvisionVerifyTimeout = int64(v)
			}
			if v, ok := timeoutsConfig["mtls_deprovision_verify_timeout"].(int); ok {
				c.MTLSDeprovisionVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["mtls_iprule_create_verify_timeout"].(int); ok {
				c.MTLSIPRuleCreateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["mtls_certificate_create_verify_timeout"].(int); ok {
				c.MTLSCertificateCreateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["mtls_certificate_delete_verify_timeout"].(int); ok {
				c.MTLSCertificateDeleteVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["kafka_cg_create_verify_timeout"].(int); ok {
				c.KafkaCGCreateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["kafka_cg_delete_verify_timeout"].(int); ok {
				c.KafkaCGDeleteVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["kafka_topic_create_verify_timeout"].(int); ok {
				c.KafkaTopicCreateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["kafka_topic_update_verify_timeout"].(int); ok {
				c.KafkaTopicUpdateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["privatelink_create_verify_timeout"].(int); ok {
				c.PrivatelinkCreateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["privatelink_delete_verify_timeout"].(int); ok {
				c.PrivatelinkDeleteVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["privatelink_allowed_acccounts_add_verify_timeout"].(int); ok {
				c.PrivatelinkAllowedAccountsAddVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["privatelink_allowed_acccounts_remove_timeout"].(int); ok {
				c.PrivatelinkAllowedAccountsRemoveVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["data_connector_create_verify_timeout"].(int); ok {
				c.DataConnectorCreateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["data_connector_settings_update_verify_timeout"].(int); ok {
				c.DataConnectorSettingsUpdateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["data_connector_delete_verify_timeout"].(int); ok {
				c.DataConnectorDeleteVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["data_connector_status_update_verify_timeout"].(int); ok {
				c.DataConnectorStatusUpdateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["postgres_credential_pre_create_verify_timeout"].(int); ok {
				c.PostgresCredentialPreCreateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["postgres_credential_create_verify_timeout"].(int); ok {
				c.PostgresCredentialCreateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["postgres_credential_delete_verify_timeout"].(int); ok {
				c.PostgresCredentialDeleteVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["shield_private_space_create_verify_timeout"].(int); ok {
				c.PrivateSpaceCreateVerifyTimeout = int64(v)
			}

			if v, ok := timeoutsConfig["app_container_release_verify_timeout"].(int); ok {
				c.AppContainerReleaseVerifyTimeout = int64(v)
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
