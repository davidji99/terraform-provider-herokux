package herokuplus

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokuplus/api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
)

var (
	ValidDynoTypesForAutoscaling = []string{"performance-m", "performance-l", "private-s", "private-m",
		"private-l", "shield-s", "shield-m", "shield-l"}
)

func resourceHerokuplusFormationAutoscaling() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuplusFormationAutoscalingCreate,
		ReadContext:   resourceHerokuplusFormationAutoscalingRead,
		UpdateContext: resourceHerokuplusFormationAutoscalingUpdate,
		DeleteContext: resourceHerokuplusFormationAutoscalingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuplusFormationAutoscalingImport,
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"formation_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"is_active": {
				Type:     schema.TypeBool,
				Required: true,
			},

			"min_quantity": {
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntAtLeast(1),
				Required:     true,
			},

			"max_quantity": {
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntAtLeast(1),
				Required:     true,
			},

			"desired_p95_response_time": {
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntAtLeast(1),
				Required:     true,
			},

			"dyno_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ValidDynoTypesForAutoscaling, true),
			},

			"set_notification_channels": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true,
			},

			"period": {
				Type:         schema.TypeInt,
				ValidateFunc: validation.IntInSlice([]int{1, 5, 10}),
				Optional:     true,
				Default:      1,
			},

			"action_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"operation": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"quantity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceHerokuplusFormationAutoscalingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API

	// Parse te import ID for the appID and formationName
	importID, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, parseErr
	}

	appID := importID[0]
	formationName := importID[1]

	monitor, _, findErr := client.Formations.FindMonitorByName(appID, formationName)
	if findErr != nil {
		return nil, findErr
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", monitor.GetAppID(), monitor.GetProcessType(), monitor.GetID()))

	readErr := resourceHerokuplusFormationAutoscalingRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf(readErr[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuplusFormationAutoscalingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	// Get app id and formation name
	appID := getAppId(d)
	formationName := getFormationName(d)

	opts := constructAutoscalingOpts(d)

	// First, find the monitor ID. This ID isn't exposed in the UI so we are going to programmatically
	// retrieve it from the API for resource creation.
	monitor, _, findErr := client.Formations.FindMonitorByName(appID, formationName)
	if findErr != nil {
		return diag.FromErr(findErr)
	}

	monitorID := monitor.GetID()

	isSet, resp, setErr := client.Formations.SetAutoscale(appID, formationName, monitorID, opts)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	if !isSet {
		return diag.Errorf("Did not successfully set autoscaling. StatusCode: %d", resp.StatusCode)
	}

	// Set the ID to be a composite of the APP_ID, FORMATION_NAME, and MONITOR_ID
	d.SetId(fmt.Sprintf("%s:%s:%s", appID, formationName, monitorID))

	return resourceHerokuplusFormationAutoscalingRead(ctx, d, meta)
}

func resourceHerokuplusFormationAutoscalingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	resourceID, parseErr := parseCompositeID(d.Id(), 3)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	monitor, _, getErr := client.Formations.GetMonitor(resourceID[0], resourceID[1], resourceID[2])
	if getErr != nil {
		return diag.FromErr(getErr)
	}

	d.Set("app_id", monitor.GetAppID())
	d.Set("formation_name", monitor.GetProcessType())
	d.Set("is_active", monitor.GetIsActive())
	d.Set("min_quantity", monitor.GetMinQuantity())
	d.Set("max_quantity", monitor.GetMaxQuantity())
	d.Set("desired_p95_response_time", monitor.GetValue())
	d.Set("period", monitor.GetPeriod())
	d.Set("action_type", monitor.GetActionType())
	d.Set("operation", monitor.GetOperation())

	return nil
}

func resourceHerokuplusFormationAutoscalingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	// Get app id and formation name
	appID := getAppId(d)
	formationName := getFormationName(d)

	opts := constructAutoscalingOpts(d)

	isSet, resp, setErr := client.Formations.SetAutoscale(appID, formationName, d.Id(), opts)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	if !isSet {
		return diag.Errorf("Did not successfully set autoscaling. StatusCode: %d", resp.StatusCode)
	}

	return resourceHerokuplusFormationAutoscalingRead(ctx, d, meta)
}

func resourceHerokuplusFormationAutoscalingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//client := meta.(*Config).API
	//
	//resourceID, parseErr := parseCompositeID(d.Id(), 3)
	//if parseErr != nil {
	//	return diag.FromErr(parseErr)
	//}
	//
	//// Setting default values for the PATCH request to disable the autoscaling
	//opts := &api.AutoscalingRequest{IsActive: false, Period: 1, MinQuantity: 1, MaxQuantity: 2, DesiredP95RespTime: 1000}
	//
	//isSet, resp, setErr := client.Formations.SetAutoscale(resourceID[0], resourceID[1], resourceID[2], opts)
	//if setErr != nil {
	//	return diag.FromErr(setErr)
	//}
	//
	//if !isSet {
	//	return diag.Errorf("Did not successfully set autoscaling. StatusCode: %d", resp.StatusCode)
	//}

	// It is potentially too destructive to attempt to properly disable the autoscaling without access to the last known
	// configuration of the resource. So for now, this resource will simply remove itself from state.
	// It is up to the user to determine the best course of action in the Heroku UI for the autoscaling settings.
	d.SetId("")

	return nil
}

func constructAutoscalingOpts(d *schema.ResourceData) *api.AutoscalingRequest {
	opts := &api.AutoscalingRequest{}

	if v, ok := d.GetOk("is_active"); ok {
		vs := v.(bool)
		log.Printf("[DEBUG] is_active is : %v", vs)
		opts.IsActive = vs
	}

	if v, ok := d.GetOk("dyno_type"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] dyno_type is : %s", vs)
		opts.DynoSize = vs
	}

	if v, ok := d.GetOk("min_quantity"); ok {
		vs := v.(int)
		log.Printf("[DEBUG] min_quantity is : %d", vs)
		opts.MinQuantity = vs
	}

	if v, ok := d.GetOk("max_quantity"); ok {
		vs := v.(int)
		log.Printf("[DEBUG] max_quantity is : %d", vs)
		opts.MaxQuantity = vs
	}

	if v, ok := d.GetOk("desired_p95_response_time"); ok {
		vs := v.(int)
		log.Printf("[DEBUG] desired_p95_response_time is : %d", vs)
		opts.DesiredP95RespTime = vs
	}

	if v, ok := d.GetOk("set_notification_channels"); ok {
		raw := v.([]interface{})

		vs := make([]string, 0)
		for _, r := range raw {
			vs = append(vs, r.(string))
		}

		log.Printf("[DEBUG] set_notification_channels is : %v", vs)
		opts.NotificationChannels = vs
	}

	if v, ok := d.GetOk("period"); ok {
		vs := v.(int)
		log.Printf("[DEBUG] period is : %d", vs)
		opts.Period = vs
	}

	// Define default values for certain AutoscalingRequest fields based on the fact that these request fields
	// only have a single value.
	opts.ActionType = "scale"
	opts.Quantity = 4

	return opts
}
