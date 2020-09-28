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

const (
	CertExpirationDateFormat = "02 Jan 06 15:04 -0700" // store DateTime in RC822Z format.
)

func resourceHerokuxPostgresMTLSCertificate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresMTLSCertificateCreate,
		ReadContext:   resourceHerokuxPostgresMTLSCertificateRead,
		DeleteContext: resourceHerokuxPostgresMTLSCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresMTLSCertificateImport,
		},

		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"expiration_date": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"certificate_with_chain": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"cert_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHerokuxPostgresMTLSCertificateImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API
	parsedImportID, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, parseErr
	}

	dbName := parsedImportID[0]
	certID := parsedImportID[1]

	cert, _, getErr := client.Postgres.GetMTLSCert(dbName, certID)
	if getErr != nil {
		return nil, getErr
	}

	d.SetId(fmt.Sprintf("%s:%s", dbName, cert.GetID()))

	d.Set("database_name", dbName)
	d.Set("name", cert.GetName())
	d.Set("status", cert.GetStatus().ToString())
	d.Set("private_key", cert.GetPrivateKey())
	d.Set("certificate_with_chain", cert.GetCertificateWithChain())
	d.Set("expiration_date", cert.GetExpiresAt().Format(CertExpirationDateFormat))
	d.Set("cert_id", cert.GetID())

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPostgresMTLSCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	dbName := getDatabaseName(d)

	// Create MTLS certificate
	log.Printf("[DEBUG] Creating MTLS certificate on database %s", dbName)
	cert, _, createErr := client.Postgres.CreateMTLSCert(dbName)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Printf("[DEBUG] Waiting for MTLS certificate for %s to be ready", dbName)
	stateConf := &resource.StateChangeConf{
		Pending:      []string{postgres.MTLSCertStatuses.PENDING.ToString()},
		Target:       []string{postgres.MTLSCertStatuses.READY.ToString()},
		Refresh:      MTLSSCertStateRefreshFunc(client, dbName, cert.GetID()),
		Timeout:      time.Duration(config.MTLSCertificateCreateTimeout) * time.Minute,
		PollInterval: StateRefreshPollInterval,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for MTLS certificate to be ready on %s: %s", dbName, err.Error())
	}

	d.SetId(fmt.Sprintf("%s:%s", dbName, cert.GetID()))

	return resourceHerokuxPostgresMTLSCertificateRead(ctx, d, meta)
}

func resourceHerokuxPostgresMTLSCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	ids, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	cert, response, getErr := client.Postgres.GetMTLSCert(ids[0], ids[1])
	if getErr != nil {
		if response.StatusCode == 404 {
			log.Printf("[WARN] MTLS certificate for %s not found, removing from state", ids[0])
			d.SetId("")
			return nil
		}
		return diag.FromErr(getErr)
	}

	d.Set("database_name", ids[0])
	d.Set("name", cert.GetName())
	d.Set("status", cert.GetStatus().ToString())
	d.Set("private_key", cert.GetPrivateKey())
	d.Set("certificate_with_chain", cert.GetCertificateWithChain())
	d.Set("expiration_date", cert.GetExpiresAt().Format(CertExpirationDateFormat))
	d.Set("cert_id", cert.GetID())

	return nil
}

func resourceHerokuxPostgresMTLSCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	ids, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	dbName := ids[0]
	certID := ids[1]

	log.Printf("[DEBUG] Deleting MTLS certificate on database %s", d.Id())
	_, _, deleteErr := client.Postgres.DeleteMTLSCert(dbName, certID)
	if deleteErr != nil {
		return diag.FromErr(deleteErr)
	}

	log.Printf("[DEBUG] Waiting for MTLS certificate on %s to be deleted", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending:      []string{postgres.MTLSCertStatuses.DISABLING.ToString()},
		Target:       []string{postgres.MTLSCertStatuses.DISABLED.ToString()},
		Refresh:      MTLSCertificateDeletionStateRefreshFunc(client, dbName, certID),
		Timeout:      time.Duration(config.MTLSCertificateDeleteTimeout) * time.Minute,
		PollInterval: StateRefreshPollInterval,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for MTLS certificate to be deleted on %s: %s", d.Id(), err.Error())
	}

	d.SetId("")

	return nil
}

func MTLSSCertStateRefreshFunc(client *api.Client, dbName, certID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cert, _, getErr := client.Postgres.GetMTLSCert(dbName, certID)
		if getErr != nil {
			return nil, postgres.MTLSCertStatuses.UNKNOWN.ToString(), getErr
		}

		if *cert.GetStatus() == postgres.MTLSCertStatuses.PENDING {
			log.Printf("[DEBUG] Still waiting for MTLS certificate on %s to be ready", dbName)
			return cert, cert.GetStatus().ToString(), nil
		}

		return cert, cert.GetStatus().ToString(), nil
	}
}

func MTLSCertificateDeletionStateRefreshFunc(client *api.Client, dbName, certID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cert, response, getErr := client.Postgres.GetMTLSCert(dbName, certID)
		if getErr != nil {
			if response.StatusCode == 404 {
				// 404 means the MTLS certificate was deleted
				return postgres.MTLSCert{}, postgres.MTLSCertStatuses.DISABLED.ToString(), nil
			}
			// For all other statuses, return the error.
			return nil, postgres.MTLSCertStatuses.UNKNOWN.ToString(), getErr
		}

		if *cert.GetStatus() == postgres.MTLSCertStatuses.DISABLING {
			log.Printf("[DEBUG] Still waiting for MTLS certificate on %s to be deleted", dbName)
			return cert, cert.GetStatus().ToString(), nil
		}

		return nil, "", nil
	}
}
