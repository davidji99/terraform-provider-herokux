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

func resourceHerokuxPostgresDataclipTeamAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresDataclipTeamAssociationCreate,
		ReadContext:   resourceHerokuxPostgresDataclipTeamAssociationRead,
		DeleteContext: resourceHerokuxPostgresDataclipTeamAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresDataclipTeamAssociationImport,
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

			"team_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"team_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHerokuxPostgresDataclipTeamAssociationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API

	// Import id will be a composite of the dataclip slug and team name
	result, parseErr := tfph.ParseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, parseErr
	}

	slug := result[0]
	teamName := result[1]

	dataclip, _, getErr := client.Data.GetPostgresDataclip(slug)
	if getErr != nil {
		return nil, getErr
	}

	var teamShare *data.PostgresDataclipTeamShare

	for _, team := range dataclip.TeamShares {
		if team.SharedWith.GetName() == teamName {
			teamShare = team
			break
		}
	}

	if teamShare == nil {
		return nil, fmt.Errorf("%s not found on dataclip %s", teamName, slug)
	}

	d.SetId(teamShare.GetID())
	d.Set("dataclip_id", teamShare.GetClipID())
	d.Set("dataclip_slug", slug)
	d.Set("team_id", teamShare.SharedWith.GetID())
	d.Set("team_name", teamShare.SharedWith.GetName())

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPostgresDataclipTeamAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	dataclipID := d.Get("dataclip_id").(string)
	teamID := d.Get("team_id").(string)

	log.Printf("[DEBUG] Sharing postgres dataclip %s with team %s", dataclipID, teamID)

	teamShare, _, shareErr := client.Data.SharePostgresDataclipWithTeam(dataclipID, teamID)
	if shareErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to share postgres dataclip %s with team %s", dataclipID, teamID),
			Detail:   shareErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Shared postgres dataclip %s with team %s", dataclipID, teamID)

	d.SetId(teamShare.GetID())

	return resourceHerokuxPostgresDataclipTeamAssociationRead(ctx, d, meta)
}

func resourceHerokuxPostgresDataclipTeamAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	dataclipSlug := d.Get("dataclip_slug").(string)

	// While the team share has its own UUID, there's no known GraphQL query to retrieve just the team share resource.
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

	var teamShare *data.PostgresDataclipTeamShare

	for _, team := range dataclip.TeamShares {
		if team.GetID() == d.Id() {
			teamShare = team
			break
		}
	}

	if teamShare == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to find dataclip teamshare ID %s postgres dataclip %s", d.Id(), dataclip.GetID()),
			Detail:   "this occurred during state refresh of the [herokux_postgres_dataclip_team_association] resource",
		})
		return diags
	}

	d.Set("dataclip_slug", dataclipSlug)
	d.Set("dataclip_id", teamShare.GetClipID())
	d.Set("team_id", teamShare.SharedWith.GetID())
	d.Set("team_name", teamShare.SharedWith.GetName())

	return diags
}

func resourceHerokuxPostgresDataclipTeamAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	teamName := d.Get("team_name").(string)
	slug := d.Get("dataclip_slug").(string)

	log.Printf("[DEBUG] Removing team %s from postgres dataclip %s", teamName, slug)

	_, _, unshareErr := client.Data.UnsharePostgresDataclipWithTeam(d.Get("dataclip_id").(string), d.Id())
	if unshareErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to remove %s from postgres dataclip %s", teamName, slug),
			Detail:   unshareErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Removed team %s from postgres dataclip %s", teamName, slug)

	d.SetId("")

	return diags
}
