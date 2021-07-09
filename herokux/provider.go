package herokux

import (
	"context"
	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"time"
)

const (
	// StateRefreshPollIntervalFrequency defines the polling frequency.
	StateRefreshPollIntervalFrequency = 20
)

var (
	// StateRefreshPollInterval defines the default polling interval in seconds.
	StateRefreshPollInterval = StateRefreshPollIntervalFrequency * time.Second
)

func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:     schema.TypeString,
				Optional: true,
				// Same environment variable to keep things consistent with the Heroku provider.
				DefaultFunc: schema.EnvDefaultFunc("HEROKU_API_KEY", nil),
			},

			"metrics_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKUX_METRICS_API_URL", api.DefaultMetricAPIBaseURL),
			},

			"postgres_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKUX_POSTGRES_API_URL", api.DefaultPostgresAPIBaseURL),
			},

			"data_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKUX_DATA_API_URL", api.DefaultDataAPIBaseURL),
			},

			"redis_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKUX_REDIS_API_URL", api.DefaultRedisAPIBaseURL),
			},

			"connect_central_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKUX_CONNECT_CENTRAL_API_URL", api.DefaultConnectCentralBaseURL),
			},

			"registry_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKUX_REGISTRY_API_URL", api.DefaultRegistryBaseURL),
			},

			"kolkrabbi_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKUX_KOLKRABBI_API_URL", api.DefaultKolkrabbiAPIBaseURL),
			},

			"scheduler_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKUX_SCHEDULER_API_URL", api.DefaultSchedulerAPIBaseURL),
			},

			"platform_api_url": {
				Type:     schema.TypeString,
				Optional: true,
				// Same environment variable to keep things consistent with the Heroku provider.
				DefaultFunc: schema.EnvDefaultFunc("HEROKU_API_URL", heroku.DefaultURL),
			},

			"headers": {
				Type:     schema.TypeMap,
				Elem:     schema.TypeString,
				Optional: true,
			},

			"timeouts": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mtls_provision_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultMTLSProvisionVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"mtls_deprovision_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultMTLSMTLSDeprovisionVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"mtls_iprule_create_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultMTLSIPRuleCreateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(10),
						},

						"mtls_certificate_create_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultMTLSCertificateCreateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"mtls_certificate_delete_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultMTLSCertificateDeleteVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"kafka_cg_create_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultKafkaCGCreateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"kafka_cg_delete_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultKafkaCGDeleteVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"kafka_topic_create_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultKafkaTopicCreateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(3),
						},

						"kafka_topic_update_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultKafkaTopicUpdateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(3),
						},

						"privatelink_create_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultPrivatelinkCreateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"privatelink_delete_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultPrivatelinkDeleteVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"privatelink_allowed_acccounts_add_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultPrivatelinkAllowedAccountsAddVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(2),
						},

						"privatelink_allowed_acccounts_remove_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultPrivatelinkAllowedAccountsRemoveVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(2),
						},

						"data_connector_create_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultDataConnectorCreateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(10),
						},

						"data_connector_settings_update_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultDataConnectorSettingsUpdateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(10),
						},

						"data_connector_delete_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultDataConnectorDeleteVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(3),
						},

						"data_connector_status_update_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultDataConnectorStatusUpdateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"postgres_credential_pre_create_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultPostgresCredentialPreCreateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(20),
						},

						"postgres_credential_create_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultPostgresCredentialCreateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"postgres_credential_delete_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultPostgresCredentialDeleteVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"shield_private_space_create_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultPrivateSpaceCreateVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(10),
						},

						"app_container_release_verify_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultAppContainerReleaseVerifyTimeout,
							ValidateFunc: validation.IntAtLeast(10),
						},
					},
				},
			},

			"delays": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"postgres_settings_modify_delay": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultPostgresSettingsModifyDelay,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"connect_mapping_modify_delay": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultConnectMappingModifyDelay,
							ValidateFunc: validation.IntAtLeast(5),
						},
					},
				},
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			//"herokux_connect": dataSourceHerokuxConnect(),
			"herokux_addons":                    dataSourceHerokuxAddons(),
			"herokux_kafka_mtls_iprules":        dataSourceHerokuxMTLSIPRules(),
			"herokux_postgres_mtls_certificate": dataSourceHerokuxPostgresMTLSCertificate(),
			"herokux_registry_image":            dataSourceHerokuxRegistryImage(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"herokux_app_container_release":          resourceHerokuxAppContainerRelease(),
			"herokux_app_github_integration":         resourceHerokuxAppGithubIntegration(),
			"herokux_app_webhook":                    resourceHerokuxAppWebhook(),
			"herokux_connect_mappings":               resourceHerokuxConnectMappings(),
			"herokux_data_connector":                 resourceHerokuxDataConnector(),
			"herokux_formation_alert":                resourceHerokuxFormationAlert(),
			"herokux_formation_autoscaling":          resourceHerokuxFormationAutoscaling(),
			"herokux_kafka_consumer_group":           resourceHerokuxKafkaConsumerGroup(),
			"herokux_kafka_mtls_iprule":              resourceHerokuxKafkaMTLSIPRule(),
			"herokux_kafka_topic":                    resourceHerokuxKafkaTopic(),
			"herokux_oauth_authorization":            resourceHerokuxOauthAuthorization(),
			"herokux_pipeline_ephemeral_apps_config": resourceHerokuxPipelineEphemeralAppsConfig(),
			"herokux_pipeline_github_integration":    resourceHerokuxPipelineGithubIntegration(),
			"herokux_pipeline_member":                resourceHerokuxPipelineMember(),
			"herokux_postgres_backup_schedule":       resourceHerokuxPostgresBackupSchedule(),
			"herokux_postgres_connection_pooling":    resourceHerokuxPostgresConnectionPooling(),
			"herokux_postgres_credential":            resourceHerokuxPostgresCredential(),
			"herokux_postgres_data_link":             resourceHerokuxPostgresDataLink(),
			"herokux_postgres_maintenance_window":    resourceHerokuxPostgresMaintenanceWindow(),
			"herokux_postgres_mtls":                  resourceHerokuxPostgresMTLS(),
			"herokux_postgres_mtls_certificate":      resourceHerokuxPostgresMTLSCertificate(),
			"herokux_postgres_mtls_iprule":           resourceHerokuxPostgresMTLSIPRule(),
			"herokux_postgres_settings":              resourceHerokuxPostgresSettings(),
			"herokux_privatelink":                    resourceHerokuxPrivatelink(),
			"herokux_redis_config":                   resourceHerokuxRedisConfig(),
			"herokux_redis_maintenance_window":       resourceHerokuxRedisMaintenanceWindow(),
			"herokux_scheduler_job":                  resourceHerokuxSchedulerJob(),
			"herokux_shield_private_space":           resourceHerokuxShieldPrivateSpace(),

			//"herokux_postgres":                  resourceHerokuxPostgres(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	log.Println("[INFO] Initializing HerokuX Provider")

	var diags diag.Diagnostics

	config := NewConfig()

	if applySchemaErr := config.applySchema(d); applySchemaErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to retrieve and set provider attributes",
			Detail:   applySchemaErr.Error(),
		})

		return nil, diags
	}

	if applyNetrcErr := config.applyNetrcFile(); applyNetrcErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read and apply netrc file",
			Detail:   applyNetrcErr.Error(),
		})

		return nil, diags
	}

	if token, ok := d.GetOk("api_key"); ok {
		config.token = token.(string)
	}

	if err := config.initializeAPI(); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to initialize API client",
			Detail:   err.Error(),
		})

		return nil, diags
	}

	log.Printf("[DEBUG] Herokux Provider initialized")

	return config, diags
}
