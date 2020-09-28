package herokux

import (
	"context"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	heroku "github.com/heroku/heroku-go/v5"
	"log"
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
						"mtls_provision_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"mtls_deprovision_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"mtls_iprule_create_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"mtls_certificate_create_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"mtls_certificate_delete_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"kafka_cg_create_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"kafka_cg_delete_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"kafka_topic_create_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(3),
						},

						"kafka_topic_update_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(3),
						},

						"privatelink_create_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      15,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"privatelink_delete_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      15,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"privatelink_allowed_acccounts_add_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(2),
						},

						"privatelink_allowed_acccounts_remove_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(2),
						},

						"data_connector_create_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      20,
							ValidateFunc: validation.IntAtLeast(10),
						},

						"data_connector_delete_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(3),
						},

						"data_connector_update_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"postgres_credential_create_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(5),
						},

						"postgres_credential_delete_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(5),
						},
					},
				},
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"herokux_postgres_mtls_certificate": dataSourceHerokuxPostgresMTLSCertificate(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"herokux_data_connector":            resourceHerokuxDataConnector(),
			"herokux_formation_autoscaling":     resourceHerokuxFormationAutoscaling(),
			"herokux_kafka_consumer_group":      resourceHerokuxKafkaConsumerGroup(),
			"herokux_kafka_topic":               resourceHerokuxKafkaTopic(),
			"herokux_postgres_credential":       resourceHerokuxPostgresCredential(),
			"herokux_postgres_mtls":             resourceHerokuxPostgresMTLS(),
			"herokux_postgres_mtls_certificate": resourceHerokuxPostgresMTLSCertificate(),
			"herokux_postgres_mtls_iprule":      resourceHerokuxPostgresMTLSIPRule(),
			"herokux_privatelink":               resourceHerokuxPrivatelink(),

			//"herokux_postgres":                  resourceHerokuxPostgres(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	log.Println("[INFO] Initializing Herokux Provider")

	config := NewConfig()

	if applySchemaErr := config.applySchema(d); applySchemaErr != nil {
		return nil, diag.FromErr(applySchemaErr)
	}

	if applyNetrcErr := config.applyNetrcFile(); applyNetrcErr != nil {
		return nil, diag.FromErr(applyNetrcErr)
	}

	if token, ok := d.GetOk("api_key"); ok {
		config.token = token.(string)
	}

	if err := config.initializeAPI(); err != nil {
		return nil, diag.FromErr(err)
	}

	log.Printf("[DEBUG] Herokux Provider initialized")

	return config, nil
}
