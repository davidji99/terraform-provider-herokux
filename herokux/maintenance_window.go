package herokux

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// maintenanceStart converts the legacy weekday spelling and one- or two-digit UTC hour to API arguments.
func maintenanceStart(value string) (string, string, error) {
	day, clock, ok := strings.Cut(value, " ")
	day = strings.TrimSuffix(strings.ToLower(day), "s")
	at, err := time.Parse("15:04", clock)
	if ok && err == nil && len(day) >= 2 {
		for weekday := time.Sunday; weekday <= time.Saturday; weekday++ {
			name := weekday.String()
			if strings.HasPrefix(strings.ToLower(name), day) {
				return name, at.Format("15:04"), nil
			}
		}
	}
	return "", "", fmt.Errorf("invalid maintenance window start %q", value)
}

func sameMaintenanceStart(a, b string) bool {
	day, clock, err := maintenanceStart(a)
	otherDay, otherClock, otherErr := maintenanceStart(b)
	return err == nil && otherErr == nil && day == otherDay && clock == otherClock
}

func maintenanceWindowStart(label string) (string, error) {
	start, end, ok := strings.Cut(label, " to ")
	end, utc := strings.CutSuffix(end, " UTC")
	if !ok || !utc {
		return "", fmt.Errorf("invalid UTC maintenance window %q", label)
	}
	// A range crossing midnight may also name the ending weekday.
	if _, clock, hasDay := strings.Cut(end, " "); hasDay {
		end = clock
	}
	if _, err := time.Parse("15:04", end); err != nil {
		return "", fmt.Errorf("invalid maintenance window end %q", end)
	}
	day, clock, err := maintenanceStart(start)
	if err != nil {
		return "", err
	}
	return day + "s " + clock, nil
}

func maintenanceWindowResource(idAttr, service string) *schema.Resource {
	resolve := func(ctx context.Context, cfg *Config, ref string) (string, error) {
		if ref == "" || ref == "." || ref == ".." {
			return "", fmt.Errorf("expected an add-on name or UUID")
		}
		addon, err := cfg.PlatformAPI.AddOnInfo(ctx, url.PathEscape(ref))
		if err != nil {
			return "", err
		}
		if addon.AddonService.Name != service {
			return "", fmt.Errorf("add-on %s belongs to %q, expected %s", ref, addon.AddonService.Name, service)
		}
		if _, errs := validation.IsUUID(addon.ID, "add-on UUID"); len(errs) > 0 {
			return "", fmt.Errorf("Heroku returned an invalid add-on UUID for %s", ref)
		}
		return addon.ID, nil
	}
	read := func(ctx context.Context, d *schema.ResourceData, cfg *Config, uuid string) diag.Diagnostics {
		label, err := cfg.API.Data.GetMaintenanceWindow(ctx, uuid)
		if err != nil {
			return diag.FromErr(err)
		}
		window, err := maintenanceWindowStart(label)
		if err != nil {
			return diag.FromErr(err)
		}
		if err = d.Set(idAttr, d.Id()); err != nil {
			return diag.FromErr(err)
		}
		if !sameMaintenanceStart(d.Get("window").(string), window) {
			return diag.FromErr(d.Set("window", window))
		}
		return nil
	}
	apply := func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		cfg := meta.(*Config)
		ref := d.Get(idAttr).(string)
		d.Partial(d.Id() != "")
		day, clock, err := maintenanceStart(d.Get("window").(string))
		if err != nil {
			return diag.FromErr(err)
		}
		uuid, err := resolve(ctx, cfg, ref)
		if err != nil {
			return diag.FromErr(err)
		}
		if err = cfg.API.Data.SetMaintenanceWindow(ctx, uuid, day, clock); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(ref)
		ds := read(ctx, d, cfg, uuid)
		if !ds.HasError() {
			d.Partial(false)
		}
		return ds
	}
	return &schema.Resource{
		CreateContext: apply,
		UpdateContext: apply,
		DeleteContext: schema.NoopContext,
		ReadContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
			cfg := meta.(*Config)
			uuid, err := resolve(ctx, cfg, d.Id())
			if err != nil {
				return diag.FromErr(err)
			}
			return read(ctx, d, cfg, uuid)
		},
		Importer: &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts: resourceTimeouts(),
		Schema: map[string]*schema.Schema{
			idAttr: {Type: schema.TypeString, Required: true, ForceNew: true},
			"window": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validateMaintenanceWindow,
				DiffSuppressFunc: func(_ string, a, b string, _ *schema.ResourceData) bool { return sameMaintenanceStart(a, b) },
			},
		},
	}
}
