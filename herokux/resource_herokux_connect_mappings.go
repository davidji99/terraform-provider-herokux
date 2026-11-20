package herokux

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api/connect"
	"github.com/davidji99/tfph"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"maps"
	"net/http"
	"slices"
	"time"
)

func resourceHerokuxConnectMappings() *schema.Resource {
	return &schema.Resource{
		CreateContext: applyConnectMappings,
		ReadContext:   resourceHerokuxConnectMappingsRead,
		UpdateContext: applyConnectMappings,
		DeleteContext: resourceHerokuxConnectMappingsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxConnectMappingsImport,
		},

		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			if !d.Get("ownership_recorded").(bool) {
				return d.SetNew("ownership_recorded", true)
			}
			return nil
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

			"ownership_recorded": {Type: schema.TypeBool, Computed: true},
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
	log.Printf("This is old %v", old)
	log.Printf("This is k %v", k)
	log.Printf("This is new %v", new)

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

	log.Printf("this is the result of the diff suppress comparison: %v", string(out) == string(out2))

	return string(out) == string(out2)
}

func resourceHerokuxConnectMappingsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ids, err := tfph.ParseCompositeID(d.Id(), 2)
	if err != nil {
		return nil, err
	}
	if err = d.Set("app_id", ids[0]); err != nil {
		return nil, err
	}
	if err = d.Set("connect_id", ids[1]); err != nil {
		return nil, err
	}
	d.SetId(ids[1])
	if ds := resourceHerokuxConnectMappingsRead(ctx, d, meta); ds.HasError() {
		return nil, fmt.Errorf("import Connect mappings: %s", ds[0].Summary)
	}
	return []*schema.ResourceData{d}, nil
}

func connectMappingIDs(client *connect.Connect, id string) (map[string]string, error) {
	conn, _, err := client.GetConnection(id, connect.ConnectionGetQueryParams{Deep: true})
	if err != nil {
		return nil, err
	}
	if conn == nil {
		return nil, fmt.Errorf("empty response for Connect connection %s", id)
	}
	ids := make(map[string]string, len(conn.Mappings))
	for _, m := range conn.Mappings {
		name, id := m.GetObjectName(), m.GetID()
		if name == "" || id == "" || ids[name] != "" {
			return nil, fmt.Errorf("invalid or duplicate mapping identity for %q", name)
		}
		ids[name] = id
	}
	return ids, nil
}

func configuredMappingNames(document string) (map[string]string, error) {
	var export connect.MappingsExport
	if err := json.Unmarshal([]byte(document), &export); err != nil {
		return nil, err
	}
	if export.Mappings == nil {
		return nil, fmt.Errorf("mappings must contain an array of objects")
	}
	names := make(map[string]string, len(export.Mappings))
	for _, m := range export.Mappings {
		name := m.GetObjectName()
		if _, exists := names[name]; name == "" || exists {
			return nil, fmt.Errorf("empty or duplicate mapping object_name %q", name)
		}
		names[name] = ""
	}
	return names, nil
}

func applyConnectMappings(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// An unsuccessful adoption must not persist the planned ownership marker alongside legacy IDs.
	if d.Id() != "" {
		d.Partial(true)
	}
	cfg := meta.(*Config)
	id := getConnectID(d)
	client, err := cfg.API.Connect.ForConnection(getAppID(d), id)
	if err != nil {
		return diag.FromErr(err)
	}
	document := d.Get("mappings").(string)
	desired, err := configuredMappingNames(document)
	if err != nil {
		return diag.FromErr(err)
	}
	before, err := connectMappingIDs(client, id)
	if err != nil {
		return diag.FromErr(err)
	}
	oldRecorded, _ := d.GetChange("ownership_recorded")
	owned := map[string]interface{}{}
	if oldRecorded == true {
		owned = d.Get("mapping_data").(map[string]interface{})
	}
	for name := range desired {
		if oldID, ok := owned[name]; ok && before[name] != "" && before[name] != oldID {
			return diag.Errorf("Connect mapping %q was replaced externally; verify the replacement and reimport before adopting it", name)
		}
	}

	if len(desired) > 0 {
		if _, err = client.ImportMappings(id, []byte(document)); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(id)
	if delay := time.Duration(cfg.ConnectMappingModifyDelay) * time.Second; delay > 0 {
		select {
		case <-ctx.Done():
			return diag.FromErr(ctx.Err())
		case <-time.After(delay):
		}
	}
	current, err := connectMappingIDs(client, id)
	if err != nil {
		return diag.FromErr(err)
	}
	for name := range desired {
		if current[name] == "" {
			return diag.Errorf("Connect mapping %q is not present after import", name)
		}
		if prior := before[name]; prior != "" && current[name] != prior {
			return diag.Errorf("Connect mapping %q changed identity during import", name)
		}
		desired[name] = current[name]
	}
	for name, oldID := range owned {
		if _, keep := desired[name]; !keep && oldID == current[name] {
			if err := deleteOwnedMapping(client, current[name]); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if err := setConnectMappingOwnership(d, desired); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("ownership_recorded", true); err != nil {
		return diag.FromErr(err)
	}
	ds := readConnectMappingExport(d, client, current)
	if !ds.HasError() {
		d.Partial(false)
	}
	return ds
}

func setConnectMappingOwnership(d *schema.ResourceData, owned map[string]string) error {
	names := slices.Sorted(maps.Keys(owned))
	ids := make([]string, 0, len(names))
	for _, name := range names {
		ids = append(ids, owned[name])
	}
	if err := d.Set("mapping_data", owned); err != nil {
		return err
	}
	if err := d.Set("mapping_object_names", names); err != nil {
		return err
	}
	return d.Set("mapping_ids", ids)
}

func resourceHerokuxConnectMappingsRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := meta.(*Config).API.Connect.ForConnection(getAppID(d), getConnectID(d))
	if err != nil {
		return diag.FromErr(err)
	}
	current, err := connectMappingIDs(client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if !d.Get("ownership_recorded").(bool) {
		// Older state and bare imports contain a connection-wide export with no ownership information.
		if err := setConnectMappingOwnership(d, nil); err != nil {
			return diag.FromErr(err)
		}
	}
	return readConnectMappingExport(d, client, current)
}

func readConnectMappingExport(d *schema.ResourceData, client *connect.Connect, current map[string]string) diag.Diagnostics {
	export, _, err := client.ExportMappings(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if export == nil {
		return diag.Errorf("empty Connect export for %s", d.Id())
	}
	delete(*export, "connection")
	rows, ok := (*export)["mappings"].([]interface{})
	if !ok {
		return diag.Errorf("Connect export is missing its mappings array")
	}
	if d.Get("ownership_recorded").(bool) {
		owned := d.Get("mapping_data").(map[string]interface{})
		filtered := make([]interface{}, 0, len(owned))
		for _, raw := range rows {
			row, _ := raw.(map[string]interface{})
			name, _ := row["object_name"].(string)
			if name == "" {
				return diag.Errorf("Connect export includes a mapping without an object_name")
			}
			if id, ok := owned[name]; ok && current[name] == id {
				filtered = append(filtered, raw)
			}
		}
		(*export)["mappings"] = filtered
	}
	value, err := export.ToString()
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(d.Set("mappings", value))
}

func deleteOwnedMapping(client *connect.Connect, id string) error {
	resp, err := client.DeleteMapping(id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return err
}

func resourceHerokuxConnectMappingsDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if !d.Get("ownership_recorded").(bool) {
		d.SetId("")
		return diag.Diagnostics{{Severity: diag.Warning, Summary: "Preserved Connect mappings because their ownership had not been established by a configured apply."}}
	}
	client, err := meta.(*Config).API.Connect.ForConnection(getAppID(d), getConnectID(d))
	if err != nil {
		return diag.FromErr(err)
	}
	current, err := connectMappingIDs(client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	for name, id := range d.Get("mapping_data").(map[string]interface{}) {
		if current[name] != "" && current[name] == id {
			if err := deleteOwnedMapping(client, current[name]); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	d.SetId("")
	return nil
}
