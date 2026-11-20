package herokux

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHerokuxPostgresSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresSettingsCreate,
		ReadContext:   resourceHerokuxPostgresSettingsRead,
		UpdateContext: resourceHerokuxPostgresSettingsUpdate,
		DeleteContext: resourceHerokuxPostgresSettingsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresSettingsImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"postgres_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"log_lock_waits": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"log_connections": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"log_min_duration_statement": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(-1),
			},

			"log_min_error_statement": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"error", "log", "fatal", "panic"}, false),
			},

			"track_functions": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"none", "pl", "all"}, false),
			},

			"log_statement": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					postgres.LogStatementUpdateOptions.None.ToString(),
					postgres.LogStatementUpdateOptions.DDL.ToString(),
					postgres.LogStatementUpdateOptions.Mod.ToString(),
					postgres.LogStatementUpdateOptions.All.ToString()}, false),
			},
		},
	}
}

func resourceHerokuxPostgresSettingsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())

	readErr := resourceHerokuxPostgresSettingsRead(ctx, d, meta)
	if readErr != nil {
		return nil, fmt.Errorf("unable to import postgres settings")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPostgresSettingsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API
	opts := &postgres.SettingsRequest{}
	postgresID := getPostgresID(d)

	if v, ok := d.GetOkExists("log_lock_waits"); ok {
		vb := v.(bool)
		opts.LogLockWaits = &vb
		log.Printf("[DEBUG] settings log_lock_waits is : %v", opts.LogLockWaits)
	}

	if v, ok := d.GetOkExists("log_connections"); ok {
		vb := v.(bool)
		opts.LogConnections = &vb
		log.Printf("[DEBUG] settings log_connections is : %v", opts.LogConnections)
	}

	if v, ok := d.GetOkExists("log_min_duration_statement"); ok {
		duration := v.(int)
		opts.LogMinDurationStatement = &duration
		log.Printf("[DEBUG] settings log_min_duration_statement is : %v", opts.LogMinDurationStatement)
	}

	if v, ok := d.GetOk("log_min_error_statement"); ok {
		opts.LogMinErrorStatement = v.(string)
	}
	if v, ok := d.GetOk("track_functions"); ok {
		opts.TrackFunctions = v.(string)
	}

	if v, ok := d.GetOk("log_statement"); ok {
		opts.LogStatement = v.(string)
		log.Printf("[DEBUG] settings log_statement is : %v", opts.LogStatement)
	}

	log.Printf("[DEBUG] Updating postgres settings on %s with %v", postgresID, opts)

	_, _, updateErr := client.Postgres.UpdateSettings(postgresID, opts)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Printf("[DEBUG] Sleeping for %v minute(s) after updating postgres settings on %s", config.PostgresSettingsModifyDelay, postgresID)
	time.Sleep(time.Duration(config.PostgresSettingsModifyDelay) * time.Minute)

	// Set resource ID to be the postgres ID
	d.SetId(d.Get("postgres_id").(string))

	return resourceHerokuxPostgresSettingsRead(ctx, d, meta)
}

func resourceHerokuxPostgresSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	s, _, getErr := client.Postgres.GetSettings(d.Id())
	if getErr != nil {
		return diag.FromErr(getErr)
	}

	if s == nil {
		return diag.Errorf("empty settings response for Postgres database %s", d.Id())
	}

	d.Set("postgres_id", d.Id())
	d.Set("log_lock_waits", s.GetLogLockWaits().GetValue())
	d.Set("log_connections", s.GetLogConnections().GetValue())
	d.Set("log_min_duration_statement", s.GetLogMinDurationStatement().GetValue())
	d.Set("log_statement", s.GetLogStatement().GetValue())
	if err := d.Set("log_min_error_statement", s.LogMinErrorStatement.Value); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("track_functions", s.TrackFunctions.Value); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceHerokuxPostgresSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API
	opts := &postgres.SettingsRequest{}
	postgresID := getPostgresID(d)

	if ok := d.HasChange("log_lock_waits"); ok {
		vb := d.Get("log_lock_waits").(bool)
		opts.LogLockWaits = &vb
		log.Printf("[DEBUG] settings log_lock_waits is : %v", opts.LogLockWaits)
	}

	if ok := d.HasChange("log_connections"); ok {
		vb := d.Get("log_connections").(bool)
		opts.LogConnections = &vb
		log.Printf("[DEBUG] settings log_connections is : %v", opts.LogConnections)
	}

	if ok := d.HasChange("log_min_duration_statement"); ok {
		duration := d.Get("log_min_duration_statement").(int)
		opts.LogMinDurationStatement = &duration
		log.Printf("[DEBUG] settings log_min_duration_statement is : %v", opts.LogMinDurationStatement)
	}

	if d.HasChange("log_min_error_statement") {
		opts.LogMinErrorStatement = d.Get("log_min_error_statement").(string)
	}
	if d.HasChange("track_functions") {
		opts.TrackFunctions = d.Get("track_functions").(string)
	}

	if ok := d.HasChange("log_statement"); ok {
		opts.LogStatement = d.Get("log_statement").(string)
		log.Printf("[DEBUG] settings log_statement is : %v", opts.LogStatement)
	}

	log.Printf("[DEBUG] Updating postgres settings on %s with %v", postgresID, opts)

	_, _, updateErr := client.Postgres.UpdateSettings(postgresID, opts)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Printf("[DEBUG] Sleeping for %v minute(s) after updating postgres settings on %s", config.PostgresSettingsModifyDelay, postgresID)
	time.Sleep(time.Duration(config.PostgresSettingsModifyDelay) * time.Minute)

	return resourceHerokuxPostgresSettingsRead(ctx, d, meta)
}

func resourceHerokuxPostgresSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Not possible to delete postgres setting. Existing resource will only be removed from state.")

	d.SetId("")
	return nil
}
