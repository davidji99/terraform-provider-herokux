package herokux

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHerokuxPostgresFollower() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresFollowerCreate,
		ReadContext:   resourceHerokuxPostgresFollowerRead,
		DeleteContext: resourceHerokuxPostgresFollowerDelete,

		//Importer: &schema.ResourceImporter{
		//	StateContext: resourceHerokuxPostgresFollowerImport,
		//},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"leader_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"plan": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The Heroku Postgres database that the follower db will be created under",
			},
		},
	}
}

func resourceHerokuxPostgresFollowerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	api := meta.(*Config).API
	platformAPI := meta.(*Config).PlatformAPI

	platformAPI.AddOnCreate()

	return nil
}

func resourceHerokuxPostgresFollowerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceHerokuxPostgresFollowerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}