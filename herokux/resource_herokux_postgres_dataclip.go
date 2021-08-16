package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api/data"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
)

func resourceHerokuxPostgresDataclip() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresDataclipCreate,
		ReadContext:   resourceHerokuxPostgresDataclipRead,
		UpdateContext: resourceHerokuxPostgresDataclipUpdate,
		DeleteContext: resourceHerokuxPostgresDataclipDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresDataclipImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"postgres_attachment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},

			"title": {
				Type:     schema.TypeString,
				Required: true,
			},

			"sql": {
				Type:     schema.TypeString,
				Required: true,
			},

			"enable_shareable_links": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"creator_email": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"attachment_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"addon_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"addon_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"app_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"app_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHerokuxPostgresDataclipImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API

	dataclip, _, getErr := client.Data.GetPostgresDataclip(d.Id())
	if getErr != nil {
		return nil, getErr
	}

	d.SetId(dataclip.GetID())

	setPostgresDataclipState(d, dataclip)

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPostgresDataclipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	opts := &data.PostgresDataclipCreateRequest{}

	if v, ok := d.GetOk("title"); ok {
		opts.Title = v.(string)
		log.Printf("[DEBUG] postgres dataclip title is : %v", opts.Title)
	}

	if v, ok := d.GetOk("postgres_attachment_id"); ok {
		opts.AttachmentID = v.(string)
		log.Printf("[DEBUG] postgres dataclip attachment_id is : %v", opts.AttachmentID)
	}

	if v, ok := d.GetOk("sql"); ok {
		opts.Sql = v.(string)
		log.Printf("[DEBUG] postgres dataclip sql is : %v", opts.Sql)
	}

	log.Printf("[DEBUG] Creating postgres dataclip %s", opts.Title)

	dataclip, _, createErr := client.Data.CreatePostgresDataclip(opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to create postgres dataclip %s", opts.Title),
			Detail:   createErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Created postgres dataclip %s", opts.Title)

	d.SetId(dataclip.GetID())
	d.Set("slug", dataclip.GetSlug())

	// Enable sharing if true
	sharingToggle := d.Get("enable_shareable_links").(bool)

	if sharingToggle {
		log.Printf("[DEBUG] Enabling sharing on postgres dataclip %s", d.Id())

		_, _, enableErr := client.Data.TogglePostgresDataclipSharing(d.Get("slug").(string), true)
		if enableErr != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("unable to enable sharing for postgres dataclip %s", d.Id()),
				Detail:   enableErr.Error(),
			})
			return diags
		}

		log.Printf("[DEBUG] Enabled sharing on postgres dataclip %s", d.Id())
	}

	return resourceHerokuxPostgresDataclipRead(ctx, d, meta)
}

func resourceHerokuxPostgresDataclipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	dataclip, _, getErr := client.Data.GetPostgresDataclip(d.Get("slug").(string))
	if getErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to get dataclip %s", d.Id()),
			Detail:   getErr.Error(),
		})
		return diags
	}

	setPostgresDataclipState(d, dataclip)

	return diags
}

func resourceHerokuxPostgresDataclipUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	opts := &data.PostgresDataclipUpdateRequest{}
	opts.ClipID = d.Id()
	opts.Title = d.Get("title").(string)
	opts.AttachmentID = d.Get("postgres_attachment_id").(string)
	opts.Sql = d.Get("sql").(string)

	log.Printf("[DEBUG] Updating postgres dataclip %s", d.Id())

	_, _, updateErr := client.Data.UpdatePostgresDataclip(opts)
	if updateErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to update postgres dataclip %s", d.Id()),
			Detail:   updateErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Updated postgres dataclip %s", d.Id())

	if d.HasChange("enable_shareable_links") {
		log.Printf("[DEBUG] Updating sharing on postgres dataclip %s", d.Id())

		_, _, toggleSharingErr := client.Data.TogglePostgresDataclipSharing(d.Get("slug").(string),
			d.Get("enable_shareable_links").(bool))
		if toggleSharingErr != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("unable to toggle sharing for postgres dataclip %s", d.Id()),
				Detail:   toggleSharingErr.Error(),
			})
			return diags
		}

		log.Printf("[DEBUG] Updated sharing on postgres dataclip %s", d.Id())
	}

	return resourceHerokuxPostgresDataclipRead(ctx, d, meta)
}

func resourceHerokuxPostgresDataclipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	log.Printf("[DEBUG] Deleting postgres dataclip %s", d.Id())

	_, _, deleteErr := client.Data.DeleteDataclip(d.Id())
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to delete dataclip %s", d.Id()),
			Detail:   deleteErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Deleted postgres dataclip %s", d.Id())

	d.SetId("")

	return diags
}

func setPostgresDataclipState(d *schema.ResourceData, dataclip *data.PostgresDataclip) {
	d.Set("postgres_attachment_id", dataclip.GetDatasource().GetAttachmentID())
	d.Set("title", dataclip.GetTitle())
	d.Set("sql", dataclip.Versions[0].GetSql()) // there should only ever be one 'Version'
	d.Set("slug", dataclip.GetSlug())
	d.Set("creator_email", dataclip.Creator.GetEmail())
	d.Set("attachment_name", dataclip.GetDatasource().GetAttachmentName())
	d.Set("addon_id", dataclip.GetDatasource().GetAddonID())
	d.Set("addon_name", dataclip.GetDatasource().GetAddonName())
	d.Set("app_id", dataclip.GetDatasource().GetAppID())
	d.Set("app_name", dataclip.GetDatasource().GetAppName())

	if dataclip.GetPublicSlug() != "" {
		d.Set("enable_shareable_links", true)
	} else {
		d.Set("enable_shareable_links", false)
	}
}
