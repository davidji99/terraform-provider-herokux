package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"time"
)

func resourceHerokuxPostgresMTLS() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresMTLSProvision,
		ReadContext:   resourceHerokuxPostgresMTLSRead,
		DeleteContext: resourceHerokuxPostgresMTLSDeprovision,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresMTLSImport,
		},

		Schema: map[string]*schema.Schema{
			// While it is preferable to use the UUID, the response returns the name so we need to use the name.
			"database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"app_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"enabled_by": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"certificate_authority_chain": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceHerokuxPostgresMTLSImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())

	readErr := resourceHerokuxPostgresMTLSRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf("unable to import existing MTLS configuration")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPostgresMTLSProvision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	dbName := getDatabaseName(d)

	// Enable MTLS
	log.Printf("[DEBUG] Enabling MTLS on database %s", dbName)
	newMTLS, _, enableErr := client.Postgres.ProvisionMTLS(dbName)
	if enableErr != nil {
		return diag.FromErr(enableErr)
	}

	log.Printf("[DEBUG] Waiting for MTLS configuration on %s to be operational", dbName)
	stateConf := &resource.StateChangeConf{
		Pending:      []string{postgres.MTLSConfigStatuses.PROVISIONING.ToString()},
		Target:       []string{postgres.MTLSConfigStatuses.OPERATIONAL.ToString()},
		Refresh:      MTLSSCreationStateRefreshFunc(client, dbName),
		Timeout:      time.Duration(config.MTLSProvisionTimeout) * time.Minute,
		PollInterval: 15 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for MTLS to be operational on %s: %s", dbName, err.Error())
	}

	// Set the resource ID to be the database name
	d.SetId(newMTLS.GetAddon())

	return resourceHerokuxPostgresMTLSRead(ctx, d, meta)
}

func resourceHerokuxPostgresMTLSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	mtls, response, getErr := client.Postgres.GetMTLS(d.Id())
	if getErr != nil {
		if response.StatusCode == 404 {
			log.Printf("[WARN] MTLS configuration for %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(getErr)
	}

	d.Set("database_name", mtls.GetAddon())
	d.Set("app_name", mtls.GetApp())
	d.Set("status", mtls.GetStatus().ToString())
	d.Set("enabled_by", mtls.GetAddon())
	d.Set("certificate_authority_chain", mtls.GetCertificateAuthorityChain())

	return nil
}

func resourceHerokuxPostgresMTLSDeprovision(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	log.Printf("[DEBUG] Disabling MTLS on database %s", d.Id())
	_, _, deleteErr := client.Postgres.DeprovisionMTLS(d.Id())
	if deleteErr != nil {
		return diag.FromErr(deleteErr)
	}

	log.Printf("[DEBUG] Waiting for MTLS configuration on %s to be deprovisioned", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending:      []string{postgres.MTLSConfigStatuses.DEPROVISIONING.ToString()},
		Target:       []string{postgres.MTLSConfigStatuses.DEPROVISIONED.ToString()},
		Refresh:      MTLSSDeletionStateRefreshFunc(client, d.Id()),
		Timeout:      time.Duration(config.MTLSDeprovisionTimeout) * time.Minute,
		PollInterval: 15 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for MTLS to be deprovisioned on %s: %s", d.Id(), err.Error())
	}

	d.SetId("")

	return nil
}

func MTLSSCreationStateRefreshFunc(client *api.Client, dbName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		mtlsConfig, _, getErr := client.Postgres.GetMTLS(dbName)
		if getErr != nil {
			return nil, postgres.MTLSConfigStatuses.UNKNOWN.ToString(), getErr
		}

		if *mtlsConfig.GetStatus() == postgres.MTLSConfigStatuses.PROVISIONING {
			log.Printf("[DEBUG] Still waiting for MTLS configuration on %s to be provisioned", dbName)
			return mtlsConfig, mtlsConfig.GetStatus().ToString(), nil
		}

		return mtlsConfig, mtlsConfig.GetStatus().ToString(), nil
	}
}

func MTLSSDeletionStateRefreshFunc(client *api.Client, dbName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		mtlsConfig, response, getErr := client.Postgres.GetMTLS(dbName)
		if getErr != nil {
			if response.StatusCode == 404 {
				// 404 means the MTLS configuration was deleted
				return postgres.MTLSEndpoint{}, postgres.MTLSConfigStatuses.DEPROVISIONED.ToString(), nil
			}
			// For all other statuses, return the error.
			return nil, postgres.MTLSConfigStatuses.UNKNOWN.ToString(), getErr
		}

		if *mtlsConfig.GetStatus() == postgres.MTLSConfigStatuses.DEPROVISIONING {
			log.Printf("[DEBUG] Still waiting for MTLS configuration on %s to be deprovisioned", dbName)
			return mtlsConfig, mtlsConfig.GetStatus().ToString(), nil
		}

		return nil, "", nil
	}
}
