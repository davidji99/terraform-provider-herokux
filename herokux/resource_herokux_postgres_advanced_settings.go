package herokux

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/davidji99/terraform-provider-herokux/api/postgres/advanced"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHerokuxPostgresAdvancedSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: applyAdvancedSettings, UpdateContext: applyAdvancedSettings,
		ReadContext: readAdvancedSettings, DeleteContext: schema.NoopContext,
		Importer: &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts: resourceTimeouts(),
		Schema: map[string]*schema.Schema{
			"postgres_id": advancedDatabaseIDSchema(),
			"settings": {
				Type: schema.TypeMap, Required: true, Elem: &schema.Schema{Type: schema.TypeString},
				ValidateFunc: validateAdvancedSettings,
			},
		},
	}
}

func validateAdvancedSettings(v any, _ string) ([]string, []error) {
	var errs []error
	for name, value := range v.(map[string]any) {
		if err := advanced.ValidateSetting(name, fmt.Sprint(value)); err != nil {
			errs = append(errs, err)
		}
	}
	return nil, errs
}

// Keep the configured spelling when the API returns a boolean for a CLI alias.
func managedAdvancedSettings(current map[string]string, managed map[string]any) map[string]string {
	values := make(map[string]string)
	for name, configured := range managed {
		value, ok := current[name]
		if !ok {
			continue
		}
		spelling := configured.(string)
		switch strings.ToLower(spelling) {
		case "true", "on", "1", "yes":
			if value == "true" || value == "on" {
				value = spelling
			}
		case "false", "off", "0", "no":
			if value == "false" || value == "off" {
				value = spelling
			}
		}
		values[name] = value
	}
	return values
}

func readAdvancedSettings(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if _, errs := validation.IsUUID(d.Id(), "postgres_id"); len(errs) > 0 {
		return diag.FromErr(errs[0])
	}
	values, err := meta.(*Config).API.PostgresAdvanced.GetSettings(ctx, d.Id())
	if errors.Is(err, advanced.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	// Import exposes all current settings. Subsequent refreshes track only the managed keys.
	if d.Get("postgres_id") != "" {
		values = managedAdvancedSettings(values, d.Get("settings").(map[string]any))
	}
	if err := d.Set("postgres_id", d.Id()); err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(d.Set("settings", values))
}

func applyAdvancedSettings(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Config).API.PostgresAdvanced
	db := d.Get("postgres_id").(string)
	before, after := d.GetChange("settings")
	old := before.(map[string]any)
	managed := after.(map[string]any)
	current, err := client.GetSettings(ctx, db)
	if err != nil {
		d.Partial(true)
		return diag.FromErr(err)
	}
	observed := managedAdvancedSettings(current, managed)
	changes := make(map[string]string)
	for name, value := range managed {
		if old[name] != value && observed[name] != value {
			changes[name] = value.(string)
		}
	}
	if len(changes) == 0 {
		d.SetId(db)
		return diag.FromErr(d.Set("settings", observed))
	}
	timeout := d.Timeout(schema.TimeoutUpdate)
	if d.Id() == "" {
		timeout = d.Timeout(schema.TimeoutCreate)
	}
	if err := client.UpdateSettings(ctx, db, changes); err != nil {
		d.Partial(true)
		return diag.FromErr(err)
	}
	d.SetId(db)
	// A successful PUT acknowledges the update before GET necessarily reflects it.
	err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		current, err := client.GetSettings(ctx, db)
		if err != nil {
			return retry.NonRetryableError(err)
		}
		values := managedAdvancedSettings(current, managed)
		for name, want := range changes {
			if value, ok := values[name]; !ok || value != want {
				return retry.RetryableError(fmt.Errorf("waiting for Postgres setting %q to reach %q", name, want))
			}
		}
		if err := d.Set("settings", values); err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		d.Partial(true)
	}
	return diag.FromErr(err)
}
