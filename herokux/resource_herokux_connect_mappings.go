package herokux

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"time"
)

func resourceHerokuxConnectMappings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxConnectMappingsCreate,
		ReadContext:   resourceHerokuxConnectMappingsRead,
		UpdateContext: resourceHerokuxConnectMappingsUpdate,
		DeleteContext: resourceHerokuxConnectMappingsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxConnectMappingsImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"connect_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"mappings": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: mappingsDiffSuppress,
			},

			"mapping_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},

			"mapping_object_names": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},

			"mapping_data": {
				Type:     schema.TypeMap,
				Elem:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// mappingsDiffSuppress makes sure the provider does a diff check without being affected by spacing or indentation.
func mappingsDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	log.Println(fmt.Sprintf("This is old %v", old))
	log.Println(fmt.Sprintf("This is k %v", k))
	log.Println(fmt.Sprintf("This is new %v", new))

	in := []byte(old)
	var raw map[string]interface{}
	json.Unmarshal(in, &raw)
	out, _ := json.Marshal(raw)

	in2 := []byte(new)
	var raw2 map[string]interface{}
	json.Unmarshal(in2, &raw2)
	out2, _ := json.Marshal(raw2)

	log.Println(string(out))
	log.Println(string(out2))

	log.Println(fmt.Sprintf("this is the result of the diff suppress comparison: %v", string(out) == string(out2)))

	return string(out) == string(out2)
}

func resourceHerokuxConnectMappingsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	client := config.API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, parseErr
	}

	appID := result[0]
	connectID := result[1]

	d.Set("app_id", appID)
	d.Set("connect_id", connectID)
	d.SetId(connectID)

	setupClientErr := setupConnectAPIClient(client, appID, connectID)
	if setupClientErr != nil {
		return nil, fmt.Errorf("unable to setup API client")
	}

	readErr := resourceHerokuxConnectMappingsRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf("unable to import existing connect mapping: %v", readErr[0])
	}

	return []*schema.ResourceData{d}, nil
}

func setupConnectAPIClient(client *api.Client, appID, connectID string) diag.Diagnostics {
	var diags diag.Diagnostics

	setRootURLErr := client.Connect.SetRootAPIBaseURL(appID, connectID)
	if setRootURLErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable determine and set the root API URL for your Heroku Connect connection",
			Detail:   setRootURLErr.Error(),
		})
		return diags
	}

	return diags
}

func resourceHerokuxConnectMappingsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := meta.(*Config)
	client := config.API
	appID := getAppID(d)
	connectID := getConnectID(d)

	setupClientErr := setupConnectAPIClient(client, appID, connectID)
	if setupClientErr != nil {
		return setupClientErr
	}

	var mappings []byte
	if v, ok := d.GetOk("mappings"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] mappings: %s", vs)
		mappings = []byte(vs)
	}

	log.Printf("[DEBUG] Creating mappings on connection %s", getConnectID(d))

	_, createErr := client.Connect.ImportMappings(getConnectID(d), mappings)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create mappings",
			Detail:   createErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Created mappings on connection %s", getConnectID(d))

	// Set resource ID to connect ID
	d.SetId(getConnectID(d))

	// Arbitrary sleep before reading state.
	time.Sleep(time.Duration(config.ConnectMappingModifyDelay) * time.Second)

	return resourceHerokuxConnectMappingsRead(ctx, d, meta)
}

func resourceHerokuxConnectMappingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	appID := getAppID(d)
	connectID := getConnectID(d)

	setupClientErr := setupConnectAPIClient(client, appID, connectID)
	if setupClientErr != nil {
		return setupClientErr
	}

	connection, _, getErr := client.Connect.GetConnection(d.Id(), connect.ConnectionGetQueryParams{Deep: true})
	if getErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to retrieve mappings from connection %s", getConnectID(d)),
			Detail:   getErr.Error(),
		})
		return diags
	}

	d.Set("app_id", connection.GetAppID())
	d.Set("connect_id", connection.GetAddonID())

	// Extract all mapping IDs and object names to store in computed attribute `mapping_ids` & `mapping_object_names`.
	// Set the computed attribute `mapping_data` to be a map of key=object_name and value=mapping_id
	mappingIDs := make([]string, 0, len(connection.Mappings))
	mappingObjectNames := make([]string, 0, len(connection.Mappings))
	mappingData := make(map[string]string)
	for _, m := range connection.Mappings {
		mappingIDs = append(mappingIDs, m.GetID())
		mappingObjectNames = append(mappingObjectNames, m.GetObjectName())
		mappingData[m.GetObjectName()] = m.GetID()
	}
	d.Set("mapping_ids", mappingIDs)
	d.Set("mapping_object_names", mappingObjectNames)
	d.Set("mapping_data", mappingData)

	// Set the resource's 'mapping' attribute to whatever the Export API returns
	mappingExport, _, exportErr := client.Connect.ExportMappings(d.Id())
	if exportErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to export mappings from connection %s", getConnectID(d)),
			Detail:   exportErr.Error(),
		})
		return diags
	}

	// Prior to converting the mappingExport to a string, delete the 'connection' key as it's not needed
	// and messes with diff checking.
	delete(*mappingExport, "connection")

	mappingExportStr, strConvErr := mappingExport.ToString()
	if strConvErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to convert mappings export to string format for connection %s", getConnectID(d)),
			Detail:   strConvErr.Error(),
		})
		return diags
	}
	d.Set("mappings", mappingExportStr)

	return diags
}

func resourceHerokuxConnectMappingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := meta.(*Config)
	client := config.API
	appID := getAppID(d)
	connectID := getConnectID(d)

	setupClientErr := setupConnectAPIClient(client, appID, connectID)
	if setupClientErr != nil {
		return setupClientErr
	}

	var oldMapping, newMapping connect.MappingsExport
	objectNamesToDelete := make([]string, 0)

	// Get the diff to see what existing mappings were removed.
	if d.HasChange("mappings") {
		o, n := d.GetChange("mappings")

		// Get list of old and new mapping object names
		unmarshallErrO := json.Unmarshal([]byte(o.(string)), &oldMapping)
		if unmarshallErrO != nil {
			return diag.Errorf("unable to unmarshall old mappings")
		}
		log.Printf("[DEBUG] Here's the old mappings %v", oldMapping)

		unmarshallErrN := json.Unmarshal([]byte(n.(string)), &newMapping)
		if unmarshallErrN != nil {
			return diag.Errorf("unable to unmarshall old mappings")
		}
		log.Printf("[DEBUG] Here's the new mappings %v", newMapping)

		oldMappingObjNames := make([]string, 0, len(oldMapping.Mappings))
		for _, n := range oldMapping.Mappings {
			oldMappingObjNames = append(oldMappingObjNames, n.GetObjectName())
		}
		log.Printf("[DEBUG] Here's the old mapping object names %v", oldMappingObjNames)

		newMappingObjNames := make([]string, 0, len(newMapping.Mappings))
		for _, n := range newMapping.Mappings {
			newMappingObjNames = append(newMappingObjNames, n.GetObjectName())
		}
		log.Printf("[DEBUG] Here's the new mapping object names %v", newMappingObjNames)

		// Determine which object names were removed by checking which of the old mapping object names
		// no longer exist in the new mapping object name list.
		for _, n := range oldMappingObjNames {
			if !ContainsString(newMappingObjNames, n) {
				objectNamesToDelete = append(objectNamesToDelete, n)
			}
		}
		log.Printf("[DEBUG] List of object names to be removed %v", objectNamesToDelete)
	}

	// If there are objectNamesToDelete elements, construct a new slice to stare their mapping IDs for deletion
	// by doing a lookup with the computed attribute `mapping_data`. Then, delete those removed mappings first.
	if len(objectNamesToDelete) > 0 {
		log.Printf("Begin deleting removed object names/mappings")
		md := d.Get("mapping_data").(map[string]interface{})

		for _, n := range objectNamesToDelete {
			mappingID := md[n].(string)

			log.Printf("Updating connection mappings by deleting mapping %s (%s)", mappingID, n)

			_, deleteErr := client.Connect.DeleteMapping(mappingID)
			if deleteErr != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("unable to delete mapping %s (%s)", mappingID, n),
					Detail:   deleteErr.Error(),
				})
				return diags
			}

			log.Printf("Updated connection mappings by deleting mapping %s (%s)", mappingID, n)
		}
	}

	// Then 'update' the mappings by doing the same thing as the CREATE method
	log.Printf("Updated mappings on connect %s", getConnectID(d))

	_, update := client.Connect.ImportMappings(getConnectID(d), getConnectMappings(d))
	if update != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to update mappings",
			Detail:   update.Error(),
		})
		return diags
	}

	log.Printf("Updated mappings on connect %s", getConnectID(d))

	// Arbitrary sleep before reading state.
	time.Sleep(time.Duration(config.ConnectMappingModifyDelay) * time.Second)

	return resourceHerokuxConnectMappingsRead(ctx, d, meta)
}

func resourceHerokuxConnectMappingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	appID := getAppID(d)
	connectID := getConnectID(d)

	setupClientErr := setupConnectAPIClient(client, appID, connectID)
	if setupClientErr != nil {
		return setupClientErr
	}

	log.Printf("[DEBUG] Deleting all tracked mappings on connection %s", getConnectID(d))

	mappingIDs := d.Get("mapping_ids").([]interface{})

	if len(mappingIDs) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error deleting all tracked mappings",
			Detail:   fmt.Sprintf("Expected at least more than one mapping ID for deletion, got %d", len(mappingIDs)),
		})
	}

	// Delete all mappings managed by the resource by iterating through mapping_ids attribute.
	for _, i := range d.Get("mapping_ids").([]interface{}) {
		mappingID := i.(string)
		log.Printf("[DEBUG] Deleting mapping %s on connection %s", mappingID, getConnectID(d))

		_, deleteErr := client.Connect.DeleteMapping(mappingID)
		if deleteErr != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("unable to delete mapping %s on connection %s", mappingID, getConnectID(d)),
				Detail:   deleteErr.Error(),
			})
		}
		log.Printf("[DEBUG] Deleted mapping %s on connection %s", mappingID, getConnectID(d))
	}

	if diags != nil {
		return diags
	}

	d.SetId("")

	return nil
}
