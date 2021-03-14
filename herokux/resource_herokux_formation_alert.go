package herokux

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/metrics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
)

func resourceHerokuxFormationAlert() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxFormationAlertCreate,
		ReadContext:   resourceHerokuxFormationAlertRead,
		UpdateContext: resourceHerokuxFormationAlertUpdate,
		DeleteContext: resourceHerokuxFormationAlertDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxFormationAlertImport,
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

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						metrics.FormationMonitorNames.Latency.ToString(),
						metrics.FormationMonitorNames.ErrorRate.ToString(),
					}, false),
			},

			"threshold": {
				Type:     schema.TypeString,
				Required: true,
			},

			"sensitivity": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntInSlice(metrics.AlertSensitivityValues),
			},

			"is_active": {
				Type:     schema.TypeBool,
				Required: true,
			},

			"notification_channels": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"notification_channels"},
			},

			"notification_frequency": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      5,
				ValidateFunc: validation.IntInSlice(metrics.AlertReminderFrequencies),
			},

			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHerokuxFormationAlertImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API

	// Parse te import ID for the appID, processType, and name.
	importID, parseErr := parseCompositeID(d.Id(), 3)
	if parseErr != nil {
		return nil, parseErr
	}

	appID := importID[0]
	processType := importID[1]
	alertName := importID[2]

	alert, _, findErr := client.Metrics.FindMonitorByName(appID, processType,
		metrics.FormationMonitorName(alertName))
	if findErr != nil {
		return nil, findErr
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", alert.GetAppID(), alert.GetProcessType(), alert.GetID()))

	readErr := resourceHerokuxFormationAlertRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf(readErr[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}

func constructAppAlertOpts(d *schema.ResourceData, alertName string) *metrics.FormationAlertRequest {
	opts := &metrics.FormationAlertRequest{}

	opts.ActionType = metrics.FormationMonitorActionTypes.Alert
	opts.Name = metrics.FormationMonitorName(alertName)
	opts.Operation = metrics.DefaultOperationAttrVal

	opts.IsActive = d.Get("is_active").(bool)
	log.Printf("[DEBUG] %s alert is_active: %v", alertName, opts.IsActive)

	if v, ok := d.GetOk("threshold"); ok {
		opts.Threshold = json.Number(v.(string))
		log.Printf("[DEBUG] %s alert threshold: %v", alertName, opts.Threshold)
	}

	if v, ok := d.GetOk("sensitivity"); ok {
		opts.Sensitivity = v.(int)
		log.Printf("[DEBUG] %s alert sensitivity: %v", alertName, opts.Sensitivity)
	}

	if v, ok := d.GetOk("notification_frequency"); ok {
		opts.ReminderFrequency = v.(int)
		log.Printf("[DEBUG] %s alert email_reminder_frequency: %v", alertName, opts.ReminderFrequency)
	}

	notificationChannels := make([]string, 0)
	if v, ok := d.GetOk("notification_channels"); ok {
		raw := v.([]interface{})

		for _, r := range raw {
			notificationChannels = append(notificationChannels, r.(string))
		}
	}
	log.Printf("[DEBUG] %s alert notification_channels is : %v", alertName, notificationChannels)
	opts.NotificationChannels = notificationChannels

	return opts
}

func resourceHerokuxFormationAlertCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	metricsAPI := meta.(*Config).API
	platformAPI := meta.(*Config).PlatformAPI

	// Get app id and formation name
	appID := getAppID(d)
	processType := getProcessType(d)
	alertName := getName(d)

	// Check for existing alert.
	existingAlertCheckErr := checkForExistingMonitor(metricsAPI, appID, processType,
		metrics.FormationMonitorActionTypes.Alert.ToString(), alertName)
	if existingAlertCheckErr != nil {
		return existingAlertCheckErr
	}

	// Get information about the formation as that's needed for the formation alert POST request.
	formation, formationGetErr := platformAPI.FormationInfo(context.TODO(), appID, processType)
	if formationGetErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary: fmt.Sprintf("Unable to retrieve formation info alert for app [%s] process type [%s]",
				appID, processType),
			Detail: formationGetErr.Error(),
		})
		return diags
	}

	opts := constructAppAlertOpts(d, alertName)
	opts.DynoSize = formation.Size
	opts.Quantity = formation.Quantity

	log.Printf("[DEBUG] Creating %s alert for app [%s] process type [%s]", opts.Name, appID, processType)

	alert, _, createErr := metricsAPI.Metrics.CreateFormationAlert(appID, processType, opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary: fmt.Sprintf("Unable to create response time alert for app [%s] process type [%s]",
				appID, processType),
			Detail: createErr.Error(),
		})
		return diags
	}

	// Set the ID to be a composite of the APP_ID, FORMATION_NAME, and ALERT_ID
	d.SetId(fmt.Sprintf("%s:%s:%s", appID, processType, alert.GetID()))

	log.Printf("[DEBUG] Created %s alert for app [%s] process type [%s]", opts.Name, appID, processType)

	return resourceHerokuxFormationAlertRead(ctx, d, meta)
}

func resourceHerokuxFormationAlertRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	metricsAPI := meta.(*Config).API

	resourceID, parseErr := parseCompositeID(d.Id(), 3)
	if parseErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse resource ID into three parts",
			Detail:   parseErr.Error(),
		})
		return diags
	}

	appID := resourceID[0]
	processType := resourceID[1]
	alertId := resourceID[2]

	alert, _, getErr := metricsAPI.Metrics.GetMonitor(appID, processType, alertId)
	if getErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to refresh state",
			Detail:   getErr.Error(),
		})
		return diags
	}

	d.Set("name", alert.GetName().ToString())
	d.Set("app_id", alert.GetAppID())
	d.Set("process_type", alert.GetProcessType())
	d.Set("is_active", alert.GetIsActive())
	d.Set("sensitivity", alert.GetPeriod())
	d.Set("notification_frequency", alert.GetNotificationPeriod())
	d.Set("state", alert.GetState())
	d.Set("threshold", alert.GetValue().String())

	notificationChannels := make([]string, 0)
	if alert.HasNotificationChannels() {
		notificationChannels = alert.NotificationChannels
	}
	d.Set("notification_channels", notificationChannels)

	return diags
}

func resourceHerokuxFormationAlertUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	metricsAPI := meta.(*Config).API
	platformAPI := meta.(*Config).PlatformAPI

	resourceID, parseErr := parseCompositeID(d.Id(), 3)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	appID := resourceID[0]
	processType := resourceID[1]
	alertID := resourceID[2]

	// Get information about the formation as that's needed for the formation alert POST request.
	formation, formationGetErr := platformAPI.FormationInfo(context.TODO(), appID, processType)
	if formationGetErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary: fmt.Sprintf("Unable to retrieve formation info alert for app [%s] process type [%s]",
				appID, processType),
			Detail: formationGetErr.Error(),
		})
		return diags
	}

	opts := constructAppAlertOpts(d, getName(d))
	opts.DynoSize = formation.Size
	opts.Quantity = formation.Quantity

	log.Printf("[DEBUG] Updating %s alert for app [%s] process type [%s]", opts.Name, appID, processType)
	isUpdated, resp, setErr := metricsAPI.Metrics.UpdateFormationAlert(appID, processType, alertID, opts)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	if !isUpdated {
		return diag.Errorf("Did not successfully update %s alert for app [%s] process type [%s]. StatusCode: %d",
			opts.Name, appID, processType, resp.StatusCode)
	}

	log.Printf("[DEBUG] Updated %s alert for app [%s] process type [%s]", opts.Name, appID, processType)

	return resourceHerokuxFormationAlertRead(ctx, d, meta)
}

func resourceHerokuxFormationAlertDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	resourceID, parseErr := parseCompositeID(d.Id(), 3)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	appID := resourceID[0]
	processType := resourceID[1]
	alertID := resourceID[2]

	config := meta.(*Config)
	metricsAPI := config.API

	// Get current alert information
	monitor, _, getErr := metricsAPI.Metrics.GetMonitor(appID, processType, alertID)
	if getErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to retrieve alert %s's information prior to resource deletion", alertID),
			Detail:   getErr.Error(),
		})
		return diags
	}

	// In order to disable the monitor (alert or autoscale), we'll need to first retrieve the current details of the monitor.
	// Then, we create a new request to update the monitor with the current information but is_active set to `false`.
	// This is the only way to safely, programmatically disable the monitor like how the UI does it.
	opts := &metrics.FormationAlertRequest{
		IsActive:             false,
		NotificationChannels: monitor.NotificationChannels,
		ReminderFrequency:    monitor.GetNotificationPeriod(),
		Sensitivity:          monitor.GetPeriod(),
		ActionType:           metrics.FormationMonitorActionTypes.Alert,
		Operation:            metrics.DefaultOperationAttrVal,
		Name:                 metrics.FormationMonitorNames.Latency,
		Threshold:            monitor.GetValue(),
	}

	log.Printf("[DEBUG] Disabling alert for app %s, process_type %s, monitor %s", appID, processType, alertID)

	isSet, resp, setErr := metricsAPI.Metrics.UpdateFormationAlert(appID, processType, alertID, opts)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	if !isSet {
		return diag.Errorf("Did not successfully disable alert. StatusCode: %d", resp.StatusCode)
	}

	log.Printf("[DEBUG] Disabled alert for app %s, process_type %s, monitor %s", appID, processType, alertID)

	d.SetId("")

	return nil
}

func checkForExistingMonitor(client *api.Client, appID, processType, actionType, monitorName string) diag.Diagnostics {
	var diags diag.Diagnostics

	monitors, _, listErr := client.Metrics.ListMonitors(appID, processType)
	if listErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary: fmt.Sprintf("Unable to fetch formation monitors in order to check if %s %s already exists",
				actionType, monitorName),
			Detail: listErr.Error(),
		})
		return diags
	}

	for _, m := range monitors {
		if m.GetName().ToString() == monitorName {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf("Cannot create %s %s for app [%s] process type [%s]", actionType,
					monitorName, appID, processType),
				Detail: fmt.Sprintf("An existing %s %s already exists. Please import it first.",
					actionType, monitorName),
			})
			return diags
		}
	}

	return diags
}
