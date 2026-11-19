package herokux

import (
	"context"
	"errors"

	"github.com/davidji99/terraform-provider-herokux/api/postgres/advanced"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHerokuxPostgresStorageQuota() *schema.Resource {
	threshold := func() *schema.Schema {
		return &schema.Schema{Type: schema.TypeInt, Optional: true, ValidateFunc: validation.IntAtLeast(1)}
	}
	return &schema.Resource{
		CreateContext: applyStorageQuota, ReadContext: readStorageQuota, UpdateContext: applyStorageQuota,
		DeleteContext: schema.NoopContext,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"postgres_id": advancedDatabaseIDSchema(),
			"enforcement_action": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"notify", "restrict", "none"}, false),
			},
			"warning_gb":  threshold(),
			"critical_gb": threshold(),
		},
	}
}

func applyStorageQuota(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	db := d.Get("postgres_id").(string)
	threshold := func(key string) *int {
		if v, ok := d.GetOk(key); ok {
			n := v.(int)
			return &n
		}
		return nil
	}
	desired := advanced.StorageQuota{
		WarningGB: threshold("warning_gb"), CriticalGB: threshold("critical_gb"),
		EnforcementAction: d.Get("enforcement_action").(string),
	}
	if err := meta.(*Config).API.PostgresAdvanced.SetStorageQuota(ctx, db, desired); err != nil {
		d.Partial(true)
		return diag.FromErr(err)
	}
	d.SetId(db)
	return nil
}

func readStorageQuota(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if _, errs := validation.IsUUID(d.Id(), "postgres_id"); len(errs) > 0 {
		return diag.FromErr(errs[0])
	}
	quota, err := meta.(*Config).API.PostgresAdvanced.GetStorageQuota(ctx, d.Id())
	if errors.Is(err, advanced.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	for key, value := range map[string]any{"postgres_id": d.Id(), "warning_gb": quota.WarningGB, "critical_gb": quota.CriticalGB, "enforcement_action": quota.EnforcementAction} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
