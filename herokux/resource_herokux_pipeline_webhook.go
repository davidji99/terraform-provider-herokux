package herokux

import (
	"context"
	"fmt"
	"net/http"

	"github.com/davidji99/terraform-provider-herokux/api/platform"
	"github.com/davidji99/tfph"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHerokuxPipelineWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPipelineWebhook,
		ReadContext:   readPipelineWebhook,
		UpdateContext: updatePipelineWebhook,
		DeleteContext: deletePipelineWebhook,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			if !d.HasChange("secret") {
				return nil
			}
			if d.NewValueKnown("secret") && d.Get("secret") == "" {
				// Heroku only returns a generated signing secret when creating a subscription.
				return d.ForceNew("secret")
			}
			return d.SetNewComputed("signing_secret")
		},
		Schema: map[string]*schema.Schema{
			"pipeline_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"level": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"notify", "sync"}, false),
			},
			"event_types": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
			},
			"secret": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"authorization": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"pipeline_name":  {Type: schema.TypeString, Computed: true},
			"signing_secret": {Type: schema.TypeString, Computed: true, Sensitive: true},
		},
	}
}

func createPipelineWebhook(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pipelineID := d.Get("pipeline_id").(string)
	opts := platform.AppWebhookRequest{
		URL:           d.Get("url").(string),
		Level:         platform.WebhookLevel(d.Get("level").(string)),
		Secret:        d.Get("secret").(string),
		Authorization: d.Get("authorization").(string),
	}
	for _, event := range d.Get("event_types").(*schema.Set).List() {
		opts.Include = append(opts.Include, platform.WebhookEventType(event.(string)))
	}
	hook, resp, err := meta.(*Config).API.Platform.CreatePipelineWebhook(ctx, pipelineID, &opts)
	if err != nil {
		return diag.FromErr(err)
	}
	if hook.GetID() == "" {
		return diag.Errorf("Heroku returned a pipeline webhook without an ID")
	}
	d.SetId(pipelineID + ":" + hook.GetID())
	if opts.Secret == "" {
		secret := resp.Header().Get(platform.HerokuWebhookSecret)
		if secret == "" {
			return diag.Errorf("Heroku created the pipeline webhook without returning its signing secret")
		}
		if err := d.Set("signing_secret", secret); err != nil {
			return diag.FromErr(err)
		}
	}
	return setPipelineWebhook(d, hook)
}

func parsePipelineWebhookID(id string) (string, string, error) {
	parts, err := tfph.ParseCompositeID(id, 2)
	if err != nil {
		return "", "", err
	}
	if parts[0] == "" || parts[0] == "." || parts[0] == ".." {
		return "", "", fmt.Errorf("expected a pipeline name or UUID in the webhook ID")
	}
	if _, errs := validation.IsUUID(parts[1], "webhook UUID"); len(errs) != 0 {
		return "", "", errs[0]
	}
	return parts[0], parts[1], nil
}

func readPipelineWebhook(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pipeline, webhook, err := parsePipelineWebhookID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	hook, resp, err := meta.(*Config).API.Platform.GetPipelineWebhook(ctx, pipeline, webhook)
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	return setPipelineWebhook(d, hook)
}

func setPipelineWebhook(d *schema.ResourceData, hook *platform.PipelineWebhook) diag.Diagnostics {
	if hook.Pipeline.ID == "" || hook.GetID() == "" {
		return diag.Errorf("Heroku returned a webhook without its pipeline or webhook ID")
	}
	d.SetId(hook.Pipeline.ID + ":" + hook.GetID())
	for key, value := range map[string]interface{}{
		"pipeline_id":   hook.Pipeline.ID,
		"pipeline_name": hook.Pipeline.Name,
		"url":           hook.GetURL(),
		"level":         hook.GetLevel(),
		"event_types":   hook.EventTypes,
	} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func updatePipelineWebhook(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pipeline, webhook, err := parsePipelineWebhookID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	changes := make(map[string]interface{})
	for _, key := range []string{"url", "level", "secret", "authorization"} {
		if d.HasChange(key) {
			changes[key] = d.Get(key)
		}
	}
	if changes["authorization"] == "" {
		changes["authorization"] = nil
	}
	if d.HasChange("event_types") {
		changes["include"] = d.Get("event_types").(*schema.Set).List()
	}
	if len(changes) > 0 {
		_, _, err := meta.(*Config).API.Platform.UpdatePipelineWebhook(ctx, pipeline, webhook, changes)
		if err != nil {
			// Refresh cannot recover the old credentials if a PATCH fails.
			d.Partial(true)
			return diag.FromErr(err)
		}
		if d.HasChange("secret") {
			if err := d.Set("signing_secret", nil); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return readPipelineWebhook(ctx, d, meta)
}

func deletePipelineWebhook(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pipeline, webhook, err := parsePipelineWebhookID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	resp, err := meta.(*Config).API.Platform.DeletePipelineWebhook(ctx, pipeline, webhook)
	if err != nil && (resp == nil || resp.StatusCode() != http.StatusNotFound) {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
