package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api/data"
	"github.com/davidji99/tfph"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
)

func resourceHerokuxPostgresDataclipUserAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresDataclipUserAssociationCreate,
		ReadContext:   resourceHerokuxPostgresDataclipUserAssociationRead,
		DeleteContext: resourceHerokuxPostgresDataclipUserAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresDataclipUserAssociationImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"dataclip_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"dataclip_slug": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"shared_by_email": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHerokuxPostgresDataclipUserAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API

	// Import id will be a composite of the dataclip slug and user email
	result, parseErr := tfph.ParseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, parseErr
	}

	slug := result[0]
	email := result[1]

	dataclip, _, getErr := client.Data.GetPostgresDataclip(slug)
	if getErr != nil {
		return nil, getErr
	}

	var userShare *data.PostgresDataclipUserShare

	for _, user := range dataclip.UserShares {
		if user.SharedWith.GetEmail() == email {
			userShare = user
			break
		}
	}

	if userShare == nil {
		return nil, fmt.Errorf("%s not found on dataclip %s", email, slug)
	}

	d.SetId(userShare.GetID())
	d.Set("dataclip_id", userShare.GetClipID())
	d.Set("dataclip_slug", slug)
	d.Set("email", userShare.SharedWith.GetEmail())
	d.Set("shared_by_email", userShare.SharedBy.GetEmail())

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPostgresDataclipUserAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	dataclipID := d.Get("dataclip_id").(string)
	email := d.Get("email").(string)

	log.Printf("[DEBUG] Sharing postgres dataclip %s with %s", dataclipID, email)

	userShare, _, shareErr := client.Data.SharePostgresDataclipWithUser(dataclipID, email)
	if shareErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to share postgres dataclip %s with %s", dataclipID, email),
			Detail:   shareErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Shared postgres dataclip %s with %s", dataclipID, email)

	d.SetId(userShare.GetID())

	return resourceHerokuxPostgresDataclipUserAssociationRead(ctx, d, meta)
}

func resourceHerokuxPostgresDataclipUserAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	dataclipSlug := d.Get("dataclip_slug").(string)

	// While the user share has its own UUID, there's no known GraphQL query to retrieve just the user share resource.
	// So, we'll have to get it via the dataclip query.
	dataclip, _, getErr := client.Data.GetPostgresDataclip(dataclipSlug)
	if getErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to retrieve postgres dataclip %s", dataclipSlug),
			Detail:   getErr.Error(),
		})
		return diags
	}

	var userShare *data.PostgresDataclipUserShare

	for _, user := range dataclip.UserShares {
		if user.GetID() == d.Id() {
			userShare = user
			break
		}
	}

	if userShare == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to find dataclip usershare ID %s postgres dataclip %s", d.Id(), dataclip.GetID()),
			Detail:   "this occurred during state refresh of the [herokux_postgres_dataclip_user_association] resource",
		})
		return diags
	}

	d.Set("dataclip_slug", dataclipSlug)
	d.Set("dataclip_id", userShare.GetClipID())
	d.Set("email", userShare.SharedWith.GetEmail())
	d.Set("shared_by_email", userShare.SharedBy.GetEmail())

	return diags
}

func resourceHerokuxPostgresDataclipUserAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	log.Printf("[DEBUG] Removing user %s from postgres dataclip %s", d.Get("email").(string),
		d.Get("dataclip_slug").(string))

	_, _, unshareErr := client.Data.UnsharePostgresDataclipWithUser(d.Get("dataclip_id").(string), d.Id())
	if unshareErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary: fmt.Sprintf("unable to remove %s from postgres dataclip %s", d.Get("email").(string),
				d.Get("dataclip_slug").(string)),
			Detail: unshareErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Removed user %s from postgres dataclip %s", d.Get("email").(string),
		d.Get("dataclip_slug").(string))

	d.SetId("")

	return diags
}
