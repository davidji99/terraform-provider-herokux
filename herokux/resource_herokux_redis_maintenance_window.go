package herokux

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
)

func resourceHerokuxRedisMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxRedisMaintenanceWindowCreate,
		ReadContext:   resourceHerokuxRedisMaintenanceWindowRead,
		UpdateContext: resourceHerokuxRedisMaintenanceWindowUpdate,
		DeleteContext: resourceHerokuxRedisMaintenanceWindowDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxRedisMaintenanceWindowImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"redis_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"window": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateMaintenanceWindow,
			},
		},
	}
}

func resourceHerokuxRedisMaintenanceWindowImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())

	readErr := resourceHerokuxRedisMaintenanceWindowRead(ctx, d, meta)
	if readErr != nil {
		return nil, fmt.Errorf("unable to import maintenance window")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxRedisMaintenanceWindowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	modifyErr := setRedisMaintenanceWindow(ctx, d, meta)
	if modifyErr != nil {
		return modifyErr
	}

	d.SetId(d.Get("redis_id").(string))

	return resourceHerokuxRedisMaintenanceWindowRead(ctx, d, meta)
}

func resourceHerokuxRedisMaintenanceWindowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	modifyErr := setRedisMaintenanceWindow(ctx, d, meta)
	if modifyErr != nil {
		return modifyErr
	}

	return resourceHerokuxRedisMaintenanceWindowRead(ctx, d, meta)
}

func setRedisMaintenanceWindow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	var redisID, window string
	if v, ok := d.GetOk("window"); ok {
		window = v.(string)
		log.Printf("[DEBUG] maintenance window is : %v", window)
	}

	if v, ok := d.GetOk("redis_id"); ok {
		redisID = v.(string)
		log.Printf("[DEBUG] maintenance redis_id is : %v", redisID)
	}

	log.Printf("[DEBUG] Setting redis maintenance window on %s", redisID)

	_, _, setErr := client.Redis.SetMaintenanceWindow(redisID, window)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	return nil
}

func resourceHerokuxRedisMaintenanceWindowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	w, _, getErr := client.Redis.GetMaintenanceWindow(d.Id())
	if getErr != nil {
		return diag.FromErr(getErr)
	}

	d.Set("redis_id", d.Id())

	// Extract the window from the string returned by the API endpoint.
	regex := regexp.MustCompile(`^Maintenance.*, window is ([A-Za-z]{2,10} \d\d?:[03]0) to.*$`)

	// If the regex finds a group match, `result` should be an array with a length of two:
	// - ["Maintenance not required, window is Tuesdays 10:30 to 14:30 UTC", "Tuesdays 10:30"]
	result := regex.FindStringSubmatch(w.GetMessage())
	if len(result) == 0 {
		return diag.Errorf("Could not properly extract the maintenance window time frame. This is likely a provider bug.")
	}

	d.Set("window", result[1])

	return nil
}

func resourceHerokuxRedisMaintenanceWindowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Not possible to delete a maintenance window. Existing resource will only be removed from state.")

	d.SetId("")
	return nil
}
