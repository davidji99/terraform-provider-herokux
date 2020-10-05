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

func resourceHerokuxPostgresMaintenanceWindow() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresMaintenanceWindowCreate,
		ReadContext:   resourceHerokuxPostgresMaintenanceWindowRead,
		UpdateContext: resourceHerokuxPostgresMaintenanceWindowUpdate,
		DeleteContext: resourceHerokuxPostgresMaintenanceWindowDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresMaintenanceWindowImport,
		},

		Schema: map[string]*schema.Schema{
			"postgres_id": {
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

func validateMaintenanceWindow(v interface{}, k string) (ws []string, errors []error) {
	name := v.(string)
	if !regexp.MustCompile(`^[A-Za-z]{2,10}s \d\d?:[03]0$`).MatchString(name) {
		errors = append(errors, fmt.Errorf("maintenance window format should be 'Days HH:MM' where where MM is 00 or 30"))
	}

	return
}

func resourceHerokuxPostgresMaintenanceWindowImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())

	readErr := resourceHerokuxPostgresMaintenanceWindowRead(ctx, d, meta)
	if readErr != nil {
		return nil, fmt.Errorf("unable to import maintenance window")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPostgresMaintenanceWindowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	modifyErr := setMaintenanceWindow(ctx, d, meta)
	if modifyErr != nil {
		return modifyErr
	}

	d.SetId(d.Get("postgres_id").(string))

	return resourceHerokuxPostgresMaintenanceWindowRead(ctx, d, meta)
}

func resourceHerokuxPostgresMaintenanceWindowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	modifyErr := setMaintenanceWindow(ctx, d, meta)
	if modifyErr != nil {
		return modifyErr
	}

	return resourceHerokuxPostgresMaintenanceWindowRead(ctx, d, meta)
}

func setMaintenanceWindow(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	var postgresID, window string
	if v, ok := d.GetOk("window"); ok {
		window = v.(string)
		log.Printf("[DEBUG] maintenance window is : %v", window)
	}

	if v, ok := d.GetOk("postgres_id"); ok {
		postgresID = v.(string)
		log.Printf("[DEBUG] maintenance postgres_id is : %v", postgresID)
	}

	log.Printf("[DEBUG] Setting postgres maintenance window on %s", postgresID)

	_, _, setErr := client.Postgres.SetMaintenanceWindow(postgresID, window)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	return nil
}

func resourceHerokuxPostgresMaintenanceWindowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	w, _, getErr := client.Postgres.GetMaintenanceWindow(d.Id())
	if getErr != nil {
		return diag.FromErr(getErr)
	}

	d.Set("postgres_id", d.Id())

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

func resourceHerokuxPostgresMaintenanceWindowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Not possible to delete a maintenance window. Existing resource will only be removed from state.")

	d.SetId("")
	return nil
}
