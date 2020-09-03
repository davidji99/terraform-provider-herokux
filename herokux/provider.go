package herokux

import (
	"context"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
)

func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKU_API_KEY", nil),
			},

			"metrics_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKU_METRICS_API_URL", api.DefaultMetricAPIBaseURL),
			},

			"postgres_api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("HEROKU_POSTGRES_API_URL", api.DefaultPostgresAPIBaseURL),
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
							ValidateFunc: validation.IntAtLeast(1),
						},
						"mtls_deprovision_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      10,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
		},

		DataSourcesMap: map[string]*schema.Resource{},

		ResourcesMap: map[string]*schema.Resource{
			"herokux_formation_autoscaling": resourceHerokuxFormationAutoscaling(),
			"herokux_postgres_mtls":         resourceHerokuxPostgresMTLS(),
		},

		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	log.Println("[INFO] Initializing Herokux Provider")

	config := NewConfig()

	if token, ok := d.GetOk("api_key"); ok {
		config.token = token.(string)
	}

	if applySchemaErr := config.applySchema(d); applySchemaErr != nil {
		return nil, diag.FromErr(applySchemaErr)
	}

	if err := config.initializeAPI(); err != nil {
		return nil, diag.FromErr(err)
	}

	log.Printf("[DEBUG] Herokux Provider initialized")

	return config, nil
}
