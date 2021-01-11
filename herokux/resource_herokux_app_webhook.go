package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api/platform"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
)

func resourceHerokuxAppWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxAppWebhookCreate,
		ReadContext:   resourceHerokuxAppWebhookRead,
		UpdateContext: resourceHerokuxAppWebhookUpdate,
		DeleteContext: resourceHerokuxAppWebhookDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxAppWebhookImport,
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"level": {
				Type:     schema.TypeString,
				Required: true,
			},

			"url": {
				Type:     schema.TypeString,
				Required: true,
			},

			"event_types": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateAppWebhookName,
			},

			"secret": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			//"auth_header": {
			//	Type: schema.TypeMap,
			//	Optional: true,
			//},

			"signing_secret": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},

			"app_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func validateAppWebhookName(v interface{}, k string) (ws []string, errors []error) {
	name := v.(string)
	if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(name) {
		errors = append(errors, fmt.Errorf("webhook name must only contain lowercase letters, numbers, and dashes"))
	}
	return
}

func resourceHerokuxAppWebhookImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, parseErr
	}

	appIDorName := result[0]
	webhookIDorName := result[1]

	aw, _, getErr := client.Platform.GetAppWebhook(appIDorName, webhookIDorName)
	if getErr != nil {
		return nil, getErr
	}

	setAppWebhook(d, aw)
	d.SetId(fmt.Sprintf("%s:%s", aw.GetApp().GetID(), aw.GetID()))

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxAppWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	opts := constructAppWebhookOpts(d)
	appID := getAppID(d)

	log.Printf("[DEBUG] Creating webhook on app %s", appID)

	w, r, createErr := client.Platform.CreateAppWebhook(appID, opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create app webhook",
			Detail:   createErr.Error(),
		})
		return diags
	}

	// Set the resource ID to be a composite of appID and webhookID.
	d.SetId(fmt.Sprintf("%s:%s", appID, w.GetID()))

	log.Printf("[DEBUG] Created webhook on app %s", appID)

	// Only set the signing_secret is the `secret` attribute is not set.
	if _, ok := d.GetOk("secret"); !ok {
		setErr := d.Set("signing_secret", r.Resp.Header().Get(platform.HerokuWebhookSecret))
		if setErr != nil {
			diag.FromErr(setErr)
		}
	}

	return resourceHerokuxAppWebhookRead(ctx, d, meta)
}

func setAppWebhook(d *schema.ResourceData, aw *platform.AppWebhook) {
	d.Set("app_id", aw.GetApp().GetID())
	d.Set("app_name", aw.GetApp().GetName())
	d.Set("level", aw.GetLevel())
	d.Set("url", aw.GetURL())
	d.Set("name", aw.GetName())
	d.Set("event_types", aw.EventTypes)

	// The secret value is not returned by the API so just setting it to whatever is in the config.
	if _, ok := d.GetOk("secret"); ok {
		d.Set("secret", d.Get("secret").(string))
	}
}

func resourceHerokuxAppWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	appID := result[0]
	webhookID := result[1]

	aw, _, getErr := client.Platform.GetAppWebhook(appID, webhookID)
	if getErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to retrieve app webhook",
			Detail:   getErr.Error(),
		})
		return diags
	}

	setAppWebhook(d, aw)

	return diags
}

func resourceHerokuxAppWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	opts := constructAppWebhookOpts(d)
	appID := getAppID(d)

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	webhookID := result[1]

	log.Printf("[DEBUG] Updating webhook %s on app %s", webhookID, appID)

	_, _, updateErr := client.Platform.UpdateAppWebhook(appID, webhookID, opts)
	if updateErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to update app webhook",
			Detail:   updateErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Updated webhook %s on app %s", webhookID, appID)

	return resourceHerokuxAppWebhookRead(ctx, d, meta)
}

func resourceHerokuxAppWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	appID := result[0]
	webhookID := result[1]

	log.Printf("[DEBUG] Deleting webhook %s on app %s", webhookID, appID)

	_, deleteErr := client.Platform.DeleteAppWebhook(appID, webhookID)
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete app webhook",
			Detail:   deleteErr.Error(),
		})
		return diags
	}

	d.SetId("")

	log.Printf("[DEBUG] Deleted webhook %s on app %s", webhookID, appID)

	return nil
}

func constructAppWebhookOpts(d *schema.ResourceData) *platform.AppWebhookRequest {
	opts := &platform.AppWebhookRequest{}

	if v, ok := d.GetOk("level"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] webhook level is : %v", vs)
		opts.Level = platform.WebhookLevel(vs)
	}

	if v, ok := d.GetOk("url"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] webhook url is : %v", vs)
		opts.URL = vs
	}

	if v, ok := d.GetOk("event_types"); ok {
		eventTypes := make([]platform.WebhookEventType, 0)
		vl := v.(*schema.Set).List()
		for _, l := range vl {
			eventTypes = append(eventTypes, platform.WebhookEventType(l.(string)))
		}

		log.Printf("[DEBUG] webhook event_types is : %v", eventTypes)
		opts.Include = eventTypes
	}

	if v, ok := d.GetOk("name"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] webhook name is : %v", vs)
		opts.Name = vs
	}

	if v, ok := d.GetOk("secret"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] webhook secret is : %v", vs)
		opts.Secret = vs
	}

	return opts
}
