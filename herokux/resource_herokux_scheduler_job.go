package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api/scheduler"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
	"strconv"
)

const (
	EveryTenMinFrequency = "every_ten_minutes"
)

func resourceHerokuxSchedulerJob() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxSchedulerJobCreate,
		ReadContext:   resourceHerokuxSchedulerJobRead,
		UpdateContext: resourceHerokuxSchedulerJobUpdate,
		DeleteContext: resourceHerokuxSchedulerJobDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxSchedulerJobImport,
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"command": {
				Type:     schema.TypeString,
				Required: true,
			},

			"dyno_size": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{"Standard-1X", "Standard-2X", "Performance-M", "Performance-L"}, false),
			},

			"frequency": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^every_(ten|hour|day)_`),
					"unsupported frequency format. please refer to docs for more info."),
			},

			"last_run_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHerokuxSchedulerJobImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API
	parsedImportID, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, parseErr
	}

	appID := parsedImportID[0]
	jobID := parsedImportID[1]

	job, _, findErr := client.Scheduler.FindByID(appID, jobID)
	if findErr != nil {
		return nil, findErr
	}

	d.SetId(job.GetID())
	d.Set("app_id", appID)
	d.Set("command", job.GetAttributes().GetCommand())
	d.Set("dyno_size", job.GetAttributes().GetDynoSize())
	d.Set("last_run_at", job.GetAttributes().GetRanAt().String())

	frequency, convertErr := convertEveryAtToFrequency(job.GetAttributes().GetEvery(), job.GetAttributes().GetAt())
	if convertErr != nil {
		return nil, convertErr
	}
	d.Set("frequency", frequency)

	return []*schema.ResourceData{d}, nil
}

func constructSchedulerOpts(d *schema.ResourceData) (*scheduler.JobRequest, error) {
	opts := &scheduler.JobRequest{}

	if v, ok := d.GetOk("command"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] scheduler job command is : %v", vs)
		opts.Command = vs
	}

	if v, ok := d.GetOk("dyno_size"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] scheduler job dyno_size is : %v", vs)
		opts.DynoSize = vs
	}

	if v, ok := d.GetOk("frequency"); ok {
		frequency := v.(string)
		log.Printf("[DEBUG] scheduler job frequency is : %v", frequency)

		var parseErr error
		opts.Every, opts.At, parseErr = parseFrequency(frequency)
		if parseErr != nil {
			return nil, parseErr
		}

		log.Printf("[DEBUG] scheduler job API every value is : %v", opts.Every)
		log.Printf("[DEBUG] scheduler job API at value is : %v", opts.At)
	}

	return opts, nil
}

func resourceHerokuxSchedulerJobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	appID := getAppID(d)

	opts, optsErr := constructSchedulerOpts(d)
	if optsErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to construct schedule job opts",
			Detail:   optsErr.Error(),
		})

		return diags
	}

	log.Printf("[DEBUG] Creating scheduler job for app : %v", appID)

	job, _, createErr := client.Scheduler.Create(appID, opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to create schedule job for app %s", appID),
			Detail:   createErr.Error(),
		})

		return diags
	}

	log.Printf("[DEBUG] Created scheduler job for app : %v", appID)

	d.SetId(job.GetData().GetID())

	return resourceHerokuxSchedulerJobRead(ctx, d, meta)
}

func resourceHerokuxSchedulerJobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	appID := getAppID(d)

	job, _, findErr := client.Scheduler.FindByID(appID, d.Id())
	if findErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve schedule job %s app %s", appID, d.Id()),
			Detail:   findErr.Error(),
		})

		return diags
	}

	d.Set("app_id", appID)
	d.Set("command", job.GetAttributes().GetCommand())
	d.Set("dyno_size", job.GetAttributes().GetDynoSize())
	d.Set("last_run_at", job.GetAttributes().GetRanAt().String())

	frequency, convertErr := convertEveryAtToFrequency(job.GetAttributes().GetEvery(), job.GetAttributes().GetAt())
	if convertErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("issue retrieving value for the frequency attribute"),
			Detail:   convertErr.Error(),
		})

		return diags
	}
	d.Set("frequency", frequency)

	return nil
}

// convertEveryAtToFrequency takes the API values for `every` and `at` and converts it to the frequency attribute format.
func convertEveryAtToFrequency(every, at int) (string, error) {
	var frequency string
	switch {
	case every == 10:
		frequency = EveryTenMinFrequency
	case every == 60:
		frequency = fmt.Sprintf("every_hour_at_%d", at)
	case every == 1440:
		hour := strconv.Itoa(at / 60)
		minute := strconv.Itoa(at % 60)
		if minute == "0" {
			minute = "00"
		}
		frequency = fmt.Sprintf("every_day_at_%s:%s", hour, minute)
	default:
		return "", fmt.Errorf("unable to convert API values for `every` (%d) and `at` (%d) to frequency", every, at)
	}

	return frequency, nil
}

func resourceHerokuxSchedulerJobUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	appID := getAppID(d)

	opts, optsErr := constructSchedulerOpts(d)
	if optsErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to construct schedule job opts",
			Detail:   optsErr.Error(),
		})

		return diags
	}

	log.Printf("[DEBUG] Updating scheduler job for app : %v", appID)

	_, _, updateErr := client.Scheduler.Update(appID, d.Id(), opts)
	if updateErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to update schedule job %s for app %s", d.Id(), appID),
			Detail:   updateErr.Error(),
		})

		return diags
	}

	log.Printf("[DEBUG] Updated scheduler job for app : %v", appID)

	return resourceHerokuxSchedulerJobRead(ctx, d, meta)
}

func resourceHerokuxSchedulerJobDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	appID := getAppID(d)

	log.Printf("[DEBUG] Deleting scheduler job %s for app : %v", appID, d.Id())

	_, deleteErr := client.Scheduler.Delete(appID, d.Id())
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to delete schedule job %s for app %s", appID, d.Id()),
			Detail:   deleteErr.Error(),
		})

		return diags
	}

	log.Printf("[DEBUG] Deleted scheduler job %s for app : %v", appID, d.Id())

	d.SetId("")

	return nil
}
