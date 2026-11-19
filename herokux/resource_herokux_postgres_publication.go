package herokux

import (
	"context"
	"errors"

	"github.com/davidji99/terraform-provider-herokux/api/postgres/advanced"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHerokuxPostgresPublication() *schema.Resource {
	target := func() *schema.Schema {
		return &schema.Schema{
			Type: schema.TypeSet, Optional: true, MinItems: 1,
			Elem:         &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
			ExactlyOneOf: []string{"tables", "schemas"},
		}
	}
	return &schema.Resource{
		CreateContext: createPublication, ReadContext: readPublication,
		UpdateContext: updatePublication, DeleteContext: deletePublication,
		Importer: &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"postgres_id": advancedDatabaseIDSchema(),
			"name":        {Type: schema.TypeString, Required: true, ForceNew: true, ValidateFunc: validation.All(validation.StringIsNotEmpty, validation.StringNotInSlice([]string{".", ".."}, false))},
			"tables":      target(), "schemas": target(),
		},
	}
}

func publicationTarget(d *schema.ResourceData) advanced.PublicationTarget {
	values := func(key string) []string {
		var members []string
		for _, v := range d.Get(key).(*schema.Set).List() {
			members = append(members, v.(string))
		}
		return members
	}
	if tables := values("tables"); len(tables) > 0 {
		return advanced.PublicationTarget{Type: "tables", Tables: tables}
	}
	return advanced.PublicationTarget{Type: "schemas", Schemas: values("schemas")}
}

func createPublication(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	db := d.Get("postgres_id").(string)
	pub := advanced.Publication{Name: d.Get("name").(string), Target: publicationTarget(d)}
	if err := meta.(*Config).API.PostgresAdvanced.CreatePublication(ctx, db, pub); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(db + ":" + pub.Name)
	return nil
}

func readPublication(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	db, name, err := advancedChildID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	pub, err := meta.(*Config).API.PostgresAdvanced.GetPublication(ctx, db, name)
	if errors.Is(err, advanced.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if pub.Name != name || (pub.Target.Type != "schemas" && pub.Target.Type != "tables") {
		return diag.Errorf("unexpected publication or target type for %q: %q, %q", name, pub.Name, pub.Target.Type)
	}
	for key, value := range map[string]any{"postgres_id": db, "name": name, "schemas": pub.Target.Schemas, "tables": pub.Target.Tables} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func updatePublication(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	db, name, err := advancedChildID(d.Id())
	if err == nil {
		err = meta.(*Config).API.PostgresAdvanced.UpdatePublication(ctx, db, name, publicationTarget(d))
	}
	if err != nil {
		d.Partial(true)
		return diag.FromErr(err)
	}
	return nil
}

func deletePublication(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	db, name, err := advancedChildID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = meta.(*Config).API.PostgresAdvanced.DeletePublication(ctx, db, name)
	if err != nil && !errors.Is(err, advanced.ErrNotFound) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
