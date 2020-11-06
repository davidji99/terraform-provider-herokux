package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
)

func resourceHerokuxDataLink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxDataLinkCreate,
		ReadContext:   resourceHerokuxDataLinkRead,
		DeleteContext: resourceHerokuxDataLinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxDataLinkImport,
		},

		Schema: map[string]*schema.Schema{
			"local_db_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Description:  "The Heroku Postgres database that’s accepting the connection.",
			},

			"remote_db_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The data store that is being connected to a Heroku Postgres database.",
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateDataLinkName,
				Description:  "The name of connection between the remote and local databases.",
			},

			"link_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"remote_attachment_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func validateDataLinkName(v interface{}, k string) (ws []string, errors []error) {
	name := v.(string)
	if !regexp.MustCompile(`^[a-z][a-z0-9_]{1,61}[a-z0-9]$`).MatchString(name) {
		errors = append(errors, fmt.Errorf("link name must be between 3-63 alphanumeric characters, start with a letter, "+
			"end with an alphanumeric character, and no symbols/spaces besides an underscore"))
	}

	return
}

func resourceHerokuxDataLinkImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	_, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, fmt.Errorf("unable to import: import requires the local db ID and remote db name separated by a colon")
	}

	readErr := resourceHerokuxDataLinkRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf("unable to import data link")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxDataLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	opts := &postgres.DataLinkCreateOpts{}

	var localDB string
	if v, ok := d.GetOk("local_db_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] local_db_id is : %v", vs)
		localDB = vs
	}

	if v, ok := d.GetOk("remote_db_name"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] remote_db_name is : %v", vs)
		opts.Remote = vs
	}

	if v, ok := d.GetOk("name"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] name is : %v", vs)
		opts.Name = vs
	}

	log.Printf("[DEBUG] Creating Data Link between %s & %s", localDB, opts.Remote)

	link, _, createErr := client.Postgres.CreateDataLink(localDB, opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create new Data Link",
			Detail:   createErr.Error(),
		})
		return diags
	}

	// Set the resource ID as a composite of the local DB ID and the link name
	// as the latter is required to delete the link.
	d.SetId(fmt.Sprintf("%s:%s", localDB, link.GetName()))

	log.Printf("[DEBUG] Created Data Link between %s & %s", localDB, opts.Remote)

	return resourceHerokuxDataLinkRead(ctx, d, meta)
}

func resourceHerokuxDataLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse resource ID during state refresh (read)",
			Detail:   parseErr.Error(),
		})
		return diags
	}

	localDbID := result[0]
	linkName := result[1]

	link, response, getErr := client.Postgres.FindDataLinkByName(localDbID, linkName)
	if getErr != nil {
		if response.StatusCode == 404 {
			log.Printf("[DEBUG] Data Link %s on %s not found. Removing from state", linkName, localDbID)
			d.SetId("")

			return nil
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to find link %s for database %s", linkName, localDbID),
			Detail:   getErr.Error(),
		})
		return diags
	}

	d.Set("local_db_id", localDbID)
	d.Set("remote_db_name", link.GetRemote().GetName())
	d.Set("name", link.GetName())
	d.Set("link_id", link.GetID())
	d.Set("remote_attachment_name", link.GetRemote().GetAttachmentName())

	return diags
}

func resourceHerokuxDataLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse resource ID during deletion",
			Detail:   parseErr.Error(),
		})
		return diags
	}

	localDbID := result[0]
	linkName := result[1]

	log.Printf("[DEBUG] Deleting Data Link %s on %s", linkName, localDbID)

	_, deleteErr := client.Postgres.DeleteDataLink(localDbID, linkName)
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to delete link %s", linkName),
			Detail:   deleteErr.Error(),
		})
		return diags
	}

	d.SetId("")

	log.Printf("[DEBUG] Deleted Data Link %s on %s", linkName, localDbID)

	return diags
}
