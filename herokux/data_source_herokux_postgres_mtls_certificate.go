package herokux

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceHerokuxPostgresMTLSCertificate() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHerokuxPostgresMTLSCertRead,
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"cert_id": {
				Type:     schema.TypeString,
				Required: true,
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
		},
	}
}

func dataSourceHerokuxPostgresMTLSCertRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).API

	dbName := getDatabaseName(d)
	certID := d.Get("cert_id").(string)

	cert, _, getErr := client.Postgres.GetMTLSCert(dbName, certID)
	if getErr != nil {
		return diag.FromErr(getErr)
	}

	d.SetId(cert.GetID())

	d.Set("cert_id", cert.GetID())
	d.Set("database_name", dbName)
	d.Set("name", cert.GetName())
	d.Set("status", cert.GetStatus().ToString())
	d.Set("private_key", cert.GetPrivateKey())
	d.Set("certificate_with_chain", cert.GetCertificateWithChain())
	d.Set("expiration_date", cert.GetExpiresAt().Format(CertExpirationDateFormat))

	return nil
}
