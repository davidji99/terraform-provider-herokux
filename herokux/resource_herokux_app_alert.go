package herokux

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"time"
)

func resourceHerokuxAppAlert() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxAppAlertCreate,
		ReadContext:   resourceHerokuxAppAlertRead,
		UpdateContext: resourceHerokuxAppAlertUpdate,
		DeleteContext: resourceHerokuxAppAlertDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxAppAlertImport,
		},

		// https://www.terraform.io/docs/extend/resources/retries-and-customizable-timeouts.html
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"process_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"image_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateImageID,
			},
		},
	}
}

func resourceHerokuxAppAlertImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxAppAlertCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceHerokuxAppAlertRead(ctx, d, meta)
}

func resourceHerokuxAppAlertRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceHerokuxAppAlertUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceHerokuxAppAlertRead(ctx, d, meta)
}

func resourceHerokuxAppAlertDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
