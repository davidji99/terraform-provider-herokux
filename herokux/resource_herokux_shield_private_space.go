package herokux

import (
	"context"
	"fmt"
	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api/data"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"time"
)

var (
	herokuPrivateSpaceRegions = []string{"dublin", "frankfurt", "oregon", "sydney", "tokyo", "virginia"}
)

func resourceHerokuxShieldPrivateSpace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxShieldPrivateSpaceCreate,
		ReadContext:   resourceHerokuxShieldPrivateSpaceRead,
		UpdateContext: resourceHerokuxShieldPrivateSpaceUpdate,
		DeleteContext: resourceHerokuxShieldPrivateSpaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxShieldPrivateSpaceImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"region": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(herokuPrivateSpaceRegions, false),
			},

			"log_drain": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"team_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"cidr": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "10.0.0.0/16",
				ForceNew: true,
			},

			"data_cidr": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "10.1.0.0/16",
				ForceNew: true,
			},

			//"trusted_ip_ranges": {
			//	Type:       schema.TypeSet,
			//	Computed:   true,
			//	Optional:   true,
			//	MinItems:   0,
			//	Deprecated: "This attribute is deprecated in favor of heroku_space_inbound_ruleset. Using both at the same time will likely cause unexpected behavior.",
			//	Elem: &schema.Schema{
			//		Type: schema.TypeString,
			//	},
			//},

			"outbound_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"is_shield": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"team_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHerokuxShieldPrivateSpaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Validate to make sure the target private space is a shield one.
	config := meta.(*Config)
	client := config.PlatformAPI

	space, getErr := client.SpaceInfo(context.TODO(), d.Id())
	if getErr != nil {
		return nil, getErr
	}

	if !space.Shield {
		return nil, fmt.Errorf("Cannot import non-shield private space with this resource")
	}

	d.SetId(d.Id())

	readErr := resourceHerokuxShieldPrivateSpaceRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf("unable to import resource")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxShieldPrivateSpaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := meta.(*Config)
	client := config.PlatformAPI
	opts := heroku.SpaceCreateOpts{}

	if v, ok := d.GetOk("name"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] new shield private space name: %s", vs)
		opts.Name = vs
	}

	if v, ok := d.GetOk("region"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] new shield private space region: %s", vs)
		opts.Region = &vs
	}

	if v, ok := d.GetOk("team_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] new shield private space team_id: %s", vs)
		opts.Team = vs
	}

	if v, ok := d.GetOk("cidr"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] new shield private space cidr: %s", vs)
		opts.CIDR = &vs
	}

	if v, ok := d.GetOk("data_cidr"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] new shield private space data_cidr: %s", vs)
		opts.DataCIDR = &vs
	}

	if v, ok := d.GetOk("log_drain"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] new shield private space log_drain: %s", vs)
		opts.LogDrainURL = vs
	}

	// Always set the shield attribute to `true`
	isShield := true
	opts.Shield = &isShield

	log.Printf("[DEBUG] Creating new Shield Private Space")

	space, createErr := client.SpaceCreate(context.TODO(), opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create new Shield Private Space",
			Detail:   createErr.Error(),
		})

		return diags
	}

	log.Printf("[DEBUG] Waiting for shield private space %s to be allocated", space.ID)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"allocating"},
		Target:       []string{"Operational"},
		Refresh:      ShieldPrivateSpaceStateRefreshFunc(client, space.ID),
		Timeout:      time.Duration(config.PrivateSpaceCreateTimeout) * time.Minute,
		PollInterval: StateRefreshPollInterval,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for shield private space %s to be allocated: %s", space.ID, err.Error())
	}

	d.SetId(space.ID)

	log.Printf("[DEBUG] Created new Shield Private Space")

	return resourceHerokuxShieldPrivateSpaceRead(ctx, d, meta)
}

func ShieldPrivateSpaceStateRefreshFunc(client *heroku.Service, spaceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Check the status of both the privatelink and all allowed accounts.
		space, getErr := client.SpaceInfo(context.TODO(), spaceID)
		if getErr != nil {
			return nil, "", getErr
		}

		if space.State == "allocating" {
			log.Printf("[DEBUG] shield private space %s still allocating", spaceID)
			return space, space.State, nil
		}

		return space.State, data.PrivatelinkStatuses.OPERATIONAL.ToString(), nil
	}
}

func resourceHerokuxShieldPrivateSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := meta.(*Config)
	platformAPI := config.PlatformAPI
	api := config.API

	space, readErr := platformAPI.SpaceInfo(context.TODO(), d.Id())
	if readErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve info for shield private space %s", d.Id()),
			Detail:   readErr.Error(),
		})
		return diags
	}

	// Make sure the space is a shield private space. If not, error out.
	if !space.Shield {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Shield private space %s is not shielded", d.Id()),
			Detail:   fmt.Sprintf("This private space's `shield` attribute is %v", space.Shield),
		})
		return diags
	}

	spaceNat, natReadErr := platformAPI.SpaceNATInfo(context.TODO(), d.Id())
	if natReadErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve info for shield private space %s NAT", d.Id()),
			Detail:   natReadErr.Error(),
		})
		return diags
	}

	spaceLogDrain, _, drainGetErr := api.Platform.GetSpaceLogDrain(d.Id())
	if drainGetErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve info for shield private space %s log drain", d.Id()),
			Detail:   drainGetErr.Error(),
		})
		return diags
	}

	d.Set("name", space.Name)
	d.Set("region", space.Region.Name)
	d.Set("team_id", space.Team.ID)
	d.Set("cidr", space.CIDR)
	d.Set("data_cidr", space.DataCIDR)

	// Set computed only attributes
	d.Set("outbound_ips", spaceNat.Sources)
	d.Set("is_shield", space.Shield)
	d.Set("team_name", space.Team.Name)

	// Set log drain information
	d.Set("log_drain", spaceLogDrain.GetURL())

	return diags
}

func resourceHerokuxShieldPrivateSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := meta.(*Config)
	platformAPI := config.PlatformAPI
	api := config.API

	// Update the Space itself
	if d.HasChange("name") {
		vs := d.Get("name").(string)
		log.Printf("[DEBUG] updated shield private space name: %s", vs)
		opts := heroku.SpaceUpdateOpts{Name: &vs}

		log.Printf("[DEBUG] Updating Shield Private Space: %s", d.Id())
		_, updateErr := platformAPI.SpaceUpdate(context.TODO(), d.Id(), opts)
		if updateErr != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to update shield private space %s", d.Id()),
				Detail:   updateErr.Error(),
			})
			return diags
		}
		log.Printf("[DEBUG] Updated Shield Private Space: %s", d.Id())
	}

	// Update the Space's log drain
	if d.HasChange("log_drain") {
		url := d.Get("log_drain").(string)
		log.Printf("[DEBUG] updated shield private space log_drain: %s", url)

		log.Printf("[DEBUG] Updating Shield Private Space %s log drain", d.Id())
		_, _, updateErr := api.Platform.SetSpaceLogDrain(d.Id(), url)
		if updateErr != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to update shield private space %s's log drain", d.Id()),
				Detail:   updateErr.Error(),
			})
			return diags
		}
		log.Printf("[DEBUG] Updated Shield Private Space %s log drain", d.Id())
	}

	return resourceHerokuxShieldPrivateSpaceRead(ctx, d, meta)
}

func resourceHerokuxShieldPrivateSpaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).PlatformAPI

	log.Printf("[DEBUG] Deleting Shield Private Space: %s", d.Id())

	_, deleteErr := client.SpaceDelete(context.TODO(), d.Id())
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to delete shield private space %s", d.Id()),
			Detail:   deleteErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Deleting Shield Private Space: %s", d.Id())

	d.SetId("")

	return nil
}
