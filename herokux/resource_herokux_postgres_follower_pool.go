package herokux

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/davidji99/terraform-provider-herokux/api/postgres/advanced"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHerokuxPostgresFollowerPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: createFollowerPool,
		ReadContext:   readFollowerPool,
		UpdateContext: updateFollowerPool,
		DeleteContext: deleteFollowerPool,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts:      resourceTimeouts(),
		Schema: map[string]*schema.Schema{
			"postgres_id": advancedDatabaseIDSchema(),
			"name": {
				Type: schema.TypeString, Required: true, ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 32),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*(-[a-zA-Z0-9]+)*$`), "use letters, numbers, and single hyphens, starting with a letter and ending with a letter or number"),
					validation.StringNotInSlice([]string{"leader"}, true),
				),
			},
			"level":          {Type: schema.TypeString, Required: true, ValidateFunc: validation.StringIsNotEmpty},
			"instance_count": {Type: schema.TypeInt, Optional: true, Default: 1, ValidateFunc: validation.IntAtLeast(1)},
		},
	}
}

func createFollowerPool(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Config).API.PostgresAdvanced
	db, name := d.Get("postgres_id").(string), d.Get("name").(string)
	level, count := d.Get("level").(string), d.Get("instance_count").(int)
	pool, err := client.CreateFollowerPool(ctx, db, name, level, count)
	if err != nil {
		return diag.FromErr(err)
	}
	if pool.ID == "" {
		return diag.Errorf("Heroku created follower pool %q without returning a pool ID", name)
	}
	d.SetId(db + ":" + pool.ID)
	_, err = waitForFollowerPool(ctx, client, db, pool.ID, d.Timeout(schema.TimeoutCreate), true, func(p *advanced.Pool) bool {
		return p != nil && p.Status == "available" && p.Level == level && p.Count == count
	})
	return diag.FromErr(err)
}

func readFollowerPool(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	db, id, err := advancedChildID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	pool, err := meta.(*Config).API.PostgresAdvanced.GetFollowerPool(ctx, db, id)
	if errors.Is(err, advanced.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	return setFollowerPool(d, db, pool)
}

func setFollowerPool(d *schema.ResourceData, db string, pool *advanced.Pool) diag.Diagnostics {
	for k, v := range map[string]any{"postgres_id": db, "name": pool.Name, "level": pool.Level, "instance_count": pool.Count} {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func waitForFollowerPool(ctx context.Context, client *advanced.Client, db, id string, timeout time.Duration, mayAppear bool, done func(*advanced.Pool) bool) (*advanced.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		pool, err := client.GetFollowerPool(ctx, db, id)
		if err != nil && !errors.Is(err, advanced.ErrNotFound) {
			return nil, err
		}
		if done(pool) {
			return pool, nil
		}
		if pool == nil && !mayAppear {
			return nil, err
		}
		if pool != nil {
			mayAppear = false
		}
		select {
		case <-ctx.Done():
			return pool, fmt.Errorf("waiting for follower pool %s: %w", id, ctx.Err())
		case <-time.After(2 * time.Second):
		}
	}
}

func updateFollowerPool(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	db, id, err := advancedChildID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	client := meta.(*Config).API.PostgresAdvanced
	level, count := d.Get("level").(string), d.Get("instance_count").(int)
	timeout := d.Timeout(schema.TimeoutUpdate)
	available := func(p *advanced.Pool) bool { return p != nil && p.Status == "available" }
	pool, err := waitForFollowerPool(ctx, client, db, id, timeout, false, available)
	if err != nil {
		d.Partial(true)
		return diag.FromErr(err)
	}

	// The API resizes one dimension at a time. Compare against Heroku after each operation.
	if pool.Level != level {
		if err = client.ResizeFollowerPool(ctx, db, id, map[string]any{"level": level}); err == nil {
			pool, err = waitForFollowerPool(ctx, client, db, id, timeout, false, func(p *advanced.Pool) bool { return available(p) && p.Level == level })
		}
	}
	if err == nil && pool.Count != count {
		if err = client.ResizeFollowerPool(ctx, db, id, map[string]any{"count": count}); err == nil {
			pool, err = waitForFollowerPool(ctx, client, db, id, timeout, false, func(p *advanced.Pool) bool { return available(p) && p.Count == count })
		}
	}
	if err != nil {
		if pool == nil {
			d.Partial(true)
			return diag.FromErr(err)
		}
		return append(diag.FromErr(err), setFollowerPool(d, db, pool)...)
	}
	return setFollowerPool(d, db, pool)
}

func deleteFollowerPool(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	db, id, err := advancedChildID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	client := meta.(*Config).API.PostgresAdvanced
	_, err = client.GetFollowerPool(ctx, db, id)
	if err == nil {
		err = client.DeleteFollowerPool(ctx, db, id)
	}
	if errors.Is(err, advanced.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = waitForFollowerPool(ctx, client, db, id, d.Timeout(schema.TimeoutDelete), false, func(p *advanced.Pool) bool { return p == nil })
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
