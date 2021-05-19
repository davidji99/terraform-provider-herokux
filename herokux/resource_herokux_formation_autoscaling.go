package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api/metrics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strings"
)

func resourceHerokuxFormationAutoscaling() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxFormationAutoscalingCreate,
		ReadContext:   resourceHerokuxFormationAutoscalingRead,
		UpdateContext: resourceHerokuxFormationAutoscalingUpdate,
		DeleteContext: resourceHerokuxFormationAutoscalingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxFormationAutoscalingImport,
		},

		Timeouts: resourceTimeouts(),

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

			"dyno_size": {
				Type:      schema.TypeString,
				Required:  true,
				StateFunc: formatSize,
			},

			"notification_channels": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true,
			},

			"notification_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
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

func resourceHerokuxFormationAutoscalingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API

	// Parse te import ID for the appID and processType
	importID, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, parseErr
	}

	appID := importID[0]
	processType := importID[1]

	monitor, _, findErr := client.Metrics.FindMonitorByName(appID, processType, metrics.FormationMonitorNames.LatencyScale)
	if findErr != nil {
		return nil, findErr
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", monitor.GetAppID(), monitor.GetProcessType(), monitor.GetID()))

	readErr := resourceHerokuxFormationAutoscalingRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf(readErr[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxFormationAutoscalingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	// Get app id and formation name
	appID := getAppID(d)
	processType := getProcessType(d)

	// Check for existing autoscaling.
	existingAlertCheckErr := checkForExistingMonitor(client, appID, processType,
		metrics.FormationMonitorActionTypes.Scale.ToString(), metrics.FormationMonitorNames.LatencyScale.ToString())
	if existingAlertCheckErr != nil {
		return existingAlertCheckErr
	}

	opts := constructAutoscalingOpts(d)

	opts.Name = metrics.FormationMonitorNames.LatencyScale

	notificationChannels := make([]string, 0)
	if v, ok := d.GetOk("notification_channels"); ok {
		raw := v.([]interface{})

		for _, r := range raw {
			notificationChannels = append(notificationChannels, r.(string))
		}
	}
	log.Printf("[DEBUG] notification_channels is : %v", notificationChannels)
	opts.NotificationChannels = notificationChannels

	log.Printf("[DEBUG] Creating formation autoscaling for app %s", appID)

	fm, _, createErr := client.Metrics.CreateFormationAutoscaling(appID, processType, opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to setup autoscaling for app [%s] process type [%s]", appID, processType),
			Detail:   createErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Created formation autoscaling for app %s", appID)

	// Set the ID to be a composite of the APP_ID, FORMATION_NAME, and MONITOR_ID
	d.SetId(fmt.Sprintf("%s:%s:%s", appID, processType, fm.GetID()))

	return resourceHerokuxFormationAutoscalingRead(ctx, d, meta)
}

func resourceHerokuxFormationAutoscalingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	metricsAPI := meta.(*Config).API
	platformAPI := meta.(*Config).PlatformAPI

	resourceID, parseErr := parseCompositeID(d.Id(), 3)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	appID := resourceID[0]
	processType := resourceID[1]
	monitorID := resourceID[2]

	monitor, _, getErr := metricsAPI.Metrics.GetMonitor(appID, processType, monitorID)
	if getErr != nil {
		return diag.FromErr(getErr)
	}

	d.Set("app_id", monitor.GetAppID())
	d.Set("process_type", monitor.GetProcessType())
	d.Set("is_active", monitor.GetIsActive())
	d.Set("min_quantity", monitor.GetMinQuantity())
	d.Set("max_quantity", monitor.GetMaxQuantity())
	d.Set("action_type", monitor.GetActionType())
	d.Set("operation", monitor.GetOperation())
	d.Set("notification_period", monitor.GetNotificationPeriod())

	p95, _ := monitor.GetValue().Int64()
	d.Set("desired_p95_response_time", p95)

	// Get formation information in order to retrieve the dyno size as it's not returned by the above call.
	formation, formationGetErr := platformAPI.FormationInfo(context.TODO(), appID, processType)
	if formationGetErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to retrieve formation %s information", processType),
			Detail:   formationGetErr.Error(),
		})
		return diags
	}

	d.Set("dyno_size", formation.Size)

	notificationChannels := make([]string, 0)
	if monitor.HasNotificationChannels() {
		notificationChannels = monitor.NotificationChannels
	}
	d.Set("notification_channels", notificationChannels)

	return nil
}

func resourceHerokuxFormationAutoscalingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	// Get app id and formation name
	resourceID, parseErr := parseCompositeID(d.Id(), 3)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	appID := resourceID[0]
	processType := resourceID[1]
	monitorID := resourceID[2]

	opts := constructAutoscalingOpts(d)

	notificationChannels := make([]string, 0)
	if ok := d.HasChange("notification_channels"); ok {
		_, n := d.GetChange("notification_channels")
		if n != nil {
			raw := n.([]interface{})

			for _, r := range raw {
				notificationChannels = append(notificationChannels, r.(string))
			}
		}
	}

	log.Printf("[DEBUG] new notification_channels is : %v", notificationChannels)
	opts.NotificationChannels = notificationChannels

	log.Printf("[DEBUG] Updating formation autoscaling for app %s, formation: %s, monitor %s", appID, processType, monitorID)

	isSet, resp, setErr := client.Metrics.UpdateFormationAutoscaling(appID, processType, monitorID, opts)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	// Specific error msg if the response code is 403, which might mean the user is trying to autoscale an unsupported dyno type
	if resp.StatusCode == 403 {
		return diag.Errorf("unable to autoscale likely due to unsupported dyno type")
	}

	if !isSet {
		return diag.Errorf("Did not successfully set autoscaling. StatusCode: %d", resp.StatusCode)
	}

	log.Printf("[DEBUG] Updated formation autoscaling for app %s, formation: %s, monitor %s", appID, processType, monitorID)

	return resourceHerokuxFormationAutoscalingRead(ctx, d, meta)
}

func resourceHerokuxFormationAutoscalingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	resourceID, parseErr := parseCompositeID(d.Id(), 3)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	appID := resourceID[0]
	processType := resourceID[1]
	monitorID := resourceID[2]

	config := meta.(*Config)
	metricsAPI := config.API
	platformAPI := config.PlatformAPI

	// Get current monitor information
	monitor, _, getErr := metricsAPI.Metrics.GetMonitor(appID, processType, monitorID)
	if getErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to retrieve monitor %s's information prior to resource deletion", monitorID),
			Detail:   getErr.Error(),
		})
		return diags
	}

	// Get formation information in order to retrieve the dyno size as it's not returned by the above call.
	formation, formationGetErr := platformAPI.FormationInfo(context.TODO(), appID, processType)
	if formationGetErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to retrieve formation %s information prior to resource deletion", processType),
			Detail:   formationGetErr.Error(),
		})
		return diags
	}

	// In order to disable the autoscaling, we'll need to first retrieve the current details of the autoscaling.
	// Then, create a new request to update the autoscaling with the current information but is_active set to `false`.
	// This is the only way to safely, programmatically disable autoscaling like how the UI does it.
	p95Raw, _ := monitor.GetValue().Int64()

	opts := &metrics.FormationAutoscalingRequest{
		DynoSize:             formation.Size,
		IsActive:             false,
		MaxQuantity:          monitor.GetMaxQuantity(),
		MinQuantity:          monitor.GetMinQuantity(),
		NotificationChannels: monitor.NotificationChannels,
		NotificationPeriod:   monitor.GetNotificationPeriod(),
		DesiredP95RespTime:   int(p95Raw),
		Period:               monitor.GetPeriod(),
		ActionType:           metrics.FormationMonitorActionTypes.Scale,
		Operation:            metrics.DefaultOperationAttrVal,
		Name:                 *monitor.GetName(),
	}

	log.Printf("[DEBUG] Disabling formation autoscaling for app %s, process_type %s, monitor %s", appID, processType, monitorID)

	isSet, resp, setErr := metricsAPI.Metrics.UpdateFormationAutoscaling(appID, processType, monitorID, opts)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	if !isSet {
		return diag.Errorf("Did not successfully disable autoscaling. StatusCode: %d", resp.StatusCode)
	}

	log.Printf("[DEBUG] Disabled formation autoscaling for app %s, process_type %s, monitor %s", appID, processType, monitorID)

	d.SetId("")

	return nil
}

func constructAutoscalingOpts(d *schema.ResourceData) *metrics.FormationAutoscalingRequest {
	opts := &metrics.FormationAutoscalingRequest{}

	// Period is mainly used for app alerts so setting the value to be the default of 1.
	opts.Period = 1

	if v, ok := d.GetOk("is_active"); ok {
		vs := v.(bool)
		log.Printf("[DEBUG] is_active is : %v", vs)
		opts.IsActive = vs
	}

	if v, ok := d.GetOk("dyno_size"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] dyno_size is : %s", vs)
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

	if v, ok := d.GetOk("notification_period"); ok {
		vs := v.(int)
		log.Printf("[DEBUG] notification_period is : %d", vs)
		opts.NotificationPeriod = vs
	}

	// Define default values for certain FormationAutoscalingRequest fields based on the fact that these request fields
	// only have a single value.
	opts.ActionType = metrics.FormationMonitorActionTypes.Scale
	opts.Quantity = 1
	opts.Operation = metrics.DefaultOperationAttrVal

	return opts
}

// Guarantees a consistent format for the string that describes the
// size of a dyno. A formation's size can be "free" or "standard-1x"
// or "Private-M".
//
// Heroku's PATCH formation endpoint accepts lowercase but
// returns the capitalised version. This ensures consistent
// capitalisation for state.
//
// For all supported dyno types see:
// https://devcenter.heroku.com/articles/dyno-types
// https://devcenter.heroku.com/articles/heroku-enterprise#available-dyno-types
func formatSize(quant interface{}) string {
	if quant == nil || quant == (*string)(nil) {
		return ""
	}

	var rawQuant string
	switch t := quant.(type) {
	case string:
		log.Printf("[DEBUG] quant %v", t)
		rawQuant = quant.(string)
	case *string:
		log.Printf("[DEBUG] quant %v", t)
		rawQuant = *quant.(*string)
	default:
		return ""
	}

	// Capitalise the first descriptor, uppercase the remaining descriptors
	var formattedSlice []string
	s := strings.Split(rawQuant, "-")
	for i := range s {
		if i == 0 {
			formattedSlice = append(formattedSlice, strings.Title(s[i]))
		} else {
			formattedSlice = append(formattedSlice, strings.ToUpper(s[i]))
		}
	}

	return strings.Join(formattedSlice, "-")
}
