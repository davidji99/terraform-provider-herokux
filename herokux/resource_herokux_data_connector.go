package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"time"
)

func resourceHerokuxDataConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxDataConnectorCreate,
		ReadContext:   resourceHerokuxDataConnectorRead,
		UpdateContext: resourceHerokuxDataConnectorUpdate,
		DeleteContext: resourceHerokuxDataConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxDataConnectorImport,
		},

		Schema: map[string]*schema.Schema{
			"source_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Description:  "The UUID of the database instance whose change data you want to store",
			},

			"store_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Description:  "The UUID of the database instance that will store the change data",
			},

			"tables": {
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true,
				MinItems:    1,
				Description: "Tables to connect",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Name of the connector",
			},

			"state": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					postgres.DataConnectorStatuses.AVAILABLE.ToString(), postgres.DataConnectorStatuses.PAUSED.ToString()}, false),
			},

			"excluded_columns": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"settings": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			//"platform_version": {
			//	Type:     schema.TypeString,
			//	Optional: true,
			//	Computed: true,
			//	ForceNew: true,
			//},
		},
	}
}

func resourceHerokuxDataConnectorImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())

	readErr := resourceHerokuxDataConnectorRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf("unable to import: %s", readErr[0].Detail)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxDataConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	var sourceID, storeID string

	tables := make([]string, 0)

	if v, ok := d.GetOk("source_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] source_id is : %v", vs)
		sourceID = vs
	}

	if v, ok := d.GetOk("store_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] store_id is : %v", vs)
		storeID = vs
	}

	if v, ok := d.GetOk("tables"); ok {
		vl := v.(*schema.Set).List()
		for _, l := range vl {
			tables = append(tables, l.(string))
		}

		log.Printf("[DEBUG] tables are : %v", tables)
	}

	opts := postgres.NewDataConnectorRequest(sourceID, tables)

	if v, ok := d.GetOk("name"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] name is : %v", vs)
		opts.Name = vs
	}

	if v, ok := d.GetOk("excluded_columns"); ok {
		vl := v.(*schema.Set).List()
		excludedColumns := make([]string, 0)

		for _, l := range vl {
			excludedColumns = append(excludedColumns, l.(string))
		}
		log.Printf("[DEBUG] excluded_columns is : %v", excludedColumns)
		opts.ExcludedColumns = excludedColumns
	}

	if v, ok := d.GetOk("platform_version"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] platform_version is : %v", vs)
		opts.PlatformVersion = vs
	}

	log.Printf("[DEBUG] Creating Data Connector between %s & %s", sourceID, storeID)

	dc, _, createErr := client.Postgres.CreateDataConnector(storeID, opts)
	if createErr != nil {
		diag.FromErr(createErr)
	}

	log.Printf("[DEBUG] Waiting Data Connector %s to be provisioned", dc.GetID())

	stateConf := &resource.StateChangeConf{
		Pending:      []string{postgres.DataConnectorStatuses.CREATING.ToString()},
		Target:       []string{postgres.DataConnectorStatuses.AVAILABLE.ToString()},
		Refresh:      DataConnectorCreateStateRefreshFunc(client, dc.GetID()),
		Timeout:      time.Duration(config.DataConnectorCreateTimeout) * time.Minute,
		PollInterval: 20 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for data connector to be provisioned on %s: %s", dc.GetID(), err.Error())
	}

	log.Printf("[DEBUG] Created Data Connector %s", dc.GetID())

	d.SetId(dc.GetID())

	// Execute additional changes if needed.

	// Change the state of the data connector to paused.
	if v, ok := d.GetOk("state"); ok {
		// We only need to trigger a Pause action if the "state" attribute is set to "Paused"
		// after initial resource creation.
		if v.(string) == postgres.DataConnectorStatuses.PAUSED.ToString() {
			pauseResumeDataConnector(ctx, d, meta)
		}
	}

	// Update Data connector settings
	if _, ok := d.GetOk("settings"); ok {
		err := updateSettingsDataConnector(ctx, d, meta)
		if err.HasError() {
			return err
		}
	}

	return resourceHerokuxDataConnectorRead(ctx, d, meta)
}

func resourceHerokuxDataConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	dc, _, getErr := client.Postgres.GetDataConnector(d.Id())
	if getErr != nil {
		return diag.FromErr(getErr)
	}

	d.Set("source_id", dc.PostgresAddon.GetID())
	d.Set("store_id", dc.KafkaAddon.GetID())
	d.Set("name", dc.GetName())
	d.Set("tables", dc.Tables)
	d.Set("status", dc.Status.ToString())

	excludedColumns := make([]string, 0)
	for _, c := range dc.ExcludedColumns {
		excludedColumns = append(excludedColumns, c)
	}
	d.Set("excluded_columns", excludedColumns)

	if dc.Settings != nil {
		d.Set("settings", dc.Settings)
	} else {
		d.Set("settings", make(map[string]string, 0))
	}

	return nil
}

func resourceHerokuxDataConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange("state") {
		err := pauseResumeDataConnector(ctx, d, meta)
		if err.HasError() {
			return err
		}
	}

	if d.HasChange("settings") {
		err := updateSettingsDataConnector(ctx, d, meta)
		if err.HasError() {
			return err
		}
	}

	return resourceHerokuxDataConnectorRead(ctx, d, meta)
}

func updateSettingsDataConnector(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	settings := d.Get("settings").(map[string]interface{})

	_, _, settingsErr := client.Postgres.UpdateDataConnectorSettings(d.Id(), &postgres.DataConnectSettings{Settings: settings})
	if settingsErr != nil {
		return diag.FromErr(settingsErr)
	}

	return nil
}

func pauseResumeDataConnector(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	action := d.Get("state").(string)

	// Handle scenario if action is empty string. In this case, we don't want to do anything.
	if action == "" {
		return nil
	}

	var pendingState, targetState string

	switch action {
	case postgres.DataConnectorStatuses.PAUSED.ToString():
		log.Printf("[DEBUG] Pausing Data Connector %s", d.Id())

		_, pauseErr := client.Postgres.PauseDataConnector(d.Id())
		if pauseErr != nil {
			return diag.FromErr(pauseErr)
		}
		pendingState = postgres.DataConnectorStatuses.AVAILABLE.ToString()
		targetState = action

		log.Printf("[DEBUG] Paused Data Connector %s", d.Id())
	case postgres.DataConnectorStatuses.AVAILABLE.ToString():
		log.Printf("[DEBUG] Resuming Data Connector %s", d.Id())

		_, resumeErr := client.Postgres.ResumeDataConnector(d.Id())
		if resumeErr != nil {
			return diag.FromErr(resumeErr)
		}
		pendingState = postgres.DataConnectorStatuses.PAUSED.ToString()
		targetState = action

		log.Printf("[DEBUG] Resumed Data Connector %s", d.Id())
	default:
		return diag.Errorf("unsupported action: %s", action)
	}

	log.Printf("[DEBUG] Waiting on Data Connector %s to %s", d.Id(), action)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{pendingState},
		Target:       []string{targetState},
		Refresh:      DataConnectorStateRefreshFunc(client, d.Id(), pendingState, targetState),
		Timeout:      time.Duration(config.DataConnectorUpdateTimeout) * time.Minute,
		PollInterval: 20 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for data connector %s to %s: %s", d.Id(), action, err.Error())
	}

	return nil
}

func resourceHerokuxDataConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	log.Printf("[DEBUG] Deleting Data Connector %s", d.Id())

	_, _, deleteErr := client.Postgres.DeleteDataConnector(d.Id())
	if deleteErr != nil {
		return diag.FromErr(deleteErr)
	}

	log.Printf("[DEBUG] Waiting on Data Connector %s to be deleted", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:      []string{postgres.DataConnectorStatuses.DEPROVISIONED.ToString()},
		Target:       []string{postgres.DataConnectorStatuses.DELETED.ToString()},
		Refresh:      DataConnectorDeleteStateRefreshFunc(client, d.Id()),
		Timeout:      time.Duration(config.DataConnectorDeleteTimeout) * time.Minute,
		PollInterval: 10 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for data connector %s to be deleted: %s", d.Id(), err.Error())
	}

	log.Printf("[DEBUG] Deleted Data Connector %s", d.Id())

	d.SetId("")

	return nil
}

func DataConnectorDeleteStateRefreshFunc(client *api.Client, dcID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Check the status of the data connector.
		dc, response, getErr := client.Postgres.GetDataConnector(dcID)
		if getErr != nil {
			if response.StatusCode == 404 {
				// A 404 means the data connector has been successfully deleted
				return postgres.DataConnector{}, postgres.DataConnectorStatuses.DELETED.ToString(), nil
			}
			return nil, postgres.DataConnectorStatuses.UNKNOWN.ToString(), getErr
		}

		// Although the status is set to 'Deprovisioned', this isn't enough to indicate the data connector
		// is successfully deleted.
		if dc.Status.ToString() == postgres.DataConnectorStatuses.DEPROVISIONED.ToString() {
			log.Printf("[DEBUG] Still waiting for data connector %s to be deleted", dcID)
			return dc, dc.Status.ToString(), nil
		}

		return dc, dc.Status.ToString(), nil
	}
}

func DataConnectorCreateStateRefreshFunc(client *api.Client, dcID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Check the status of the data connector.
		dc, _, getErr := client.Postgres.GetDataConnector(dcID)
		if getErr != nil {
			return nil, postgres.DataConnectorStatuses.UNKNOWN.ToString(), getErr
		}

		if dc.Status.ToString() == postgres.DataConnectorStatuses.CREATING.ToString() {
			log.Printf("[DEBUG] Still waiting for data connector %s to be provisioned", dcID)
			return dc, dc.Status.ToString(), nil
		}

		return dc, dc.Status.ToString(), nil
	}
}

func DataConnectorStateRefreshFunc(client *api.Client, dcID string, pendingState, targetState string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Check the status of the data connector.
		dc, _, getErr := client.Postgres.GetDataConnector(dcID)
		if getErr != nil {
			return nil, postgres.DataConnectorStatuses.UNKNOWN.ToString(), getErr
		}

		if dc.Status.ToString() == pendingState {
			log.Printf("[DEBUG] Still waiting for data connector %s status to change to %s", dcID, targetState)
			return dc, dc.Status.ToString(), nil
		}

		return dc, dc.Status.ToString(), nil
	}
}
