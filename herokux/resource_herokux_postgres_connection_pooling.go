package herokux

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHerokuxPostgresConnectionPooling() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresConnectionPoolingCreate,
		ReadContext:   resourceHerokuxPostgresConnectionPoolingRead,
		DeleteContext: resourceHerokuxPostgresConnectionPoolingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresConnectionPoolingImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"postgres_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"app_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "DATABASE_CONNECTION_POOL",
				ValidateFunc: validateConnectionPoolingName,
			},

			"config_var": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func validateConnectionPoolingName(v interface{}, k string) (ws []string, errors []error) {
	name := v.(string)
	if !regexp.MustCompile(`^[A-Z][A-Z0-9_]+$`).MatchString(name) {
		errors = append(errors, fmt.Errorf("name must start with a letter and can only contain uppercase letters, numbers, and underscores"))
	}

	return
}

func resourceHerokuxPostgresConnectionPoolingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if ds := resourceHerokuxPostgresConnectionPoolingRead(ctx, d, meta); ds.HasError() {
		return nil, fmt.Errorf("import connection pool: %s", ds[0].Summary)
	}
	if d.Id() == "" {
		return nil, fmt.Errorf("connection pool attachment not found")
	}
	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPostgresConnectionPoolingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := meta.(*Config)
	api := config.API
	platformAPI := config.PlatformAPI

	postgresID := getPostgresID(d)
	appID := getAppID(d)
	opts := &postgres.ConnectionPoolingRequest{}
	opts.Credential = "default"
	opts.App = appID
	opts.Name = getName(d)

	// Check if connection pooling is available. Error out if not.
	db, _, dbInfoErr := api.Postgres.GetDB(postgresID)
	if dbInfoErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to retrieve database %s info to determine connection pooling availability", postgresID),
			Detail:   dbInfoErr.Error(),
		})
		return diags
	}

	dbInfo, infoErr := db.RetrieveSpecificInfo("Connection Pooling")
	if infoErr != nil {
		return diag.Errorf("unable to determine connection pooling")
	}

	status, err := dbInfo.FirstString()
	if err != nil {
		return diag.FromErr(err)
	}
	if status != "Available" {
		return diag.Errorf("Connection pooling not available on database %s. Status is %s", postgresID, status)
	}

	log.Printf("[DEBUG] Enabling postgres connection pooling on postgres %s", postgresID)

	attachment, _, createErr := api.Postgres.CreateConnectionPooling(postgresID, opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to create connection pooling on postgres %s", postgresID),
			Detail:   createErr.Error(),
		})
		return diags
	}

	if attachment.ID == "" {
		return diag.Errorf("Heroku returned no attachment ID after enabling pooling on %s", postgresID)
	}
	d.SetId(attachment.ID)
	wait := &resource.StateChangeConf{
		Target: []string{"attached"},
		Refresh: func() (any, string, error) {
			a, err := getPoolingAttachment(ctx, platformAPI, d.Id())
			if err != nil || a == nil {
				return nil, "", err
			}
			return a, "attached", nil
		},
		Timeout:      20 * time.Minute,
		PollInterval: 5 * time.Second,
	}
	result, err := wait.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("waiting for connection pool attachment %s: %s", d.Id(), err)
	}
	return setPoolingAttachment(d, result.(*heroku.AddOnAttachment))
}

func getPoolingAttachment(ctx context.Context, client *heroku.Service, id string) (*heroku.AddOnAttachment, error) {
	attachment, err := client.AddOnAttachmentInfo(ctx, id)
	if platformNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if attachment.ID != id || attachment.Namespace == nil || *attachment.Namespace != "connection-pooling:default" || attachment.Addon.ID == "" || attachment.App.ID == "" || attachment.Name == "" {
		return nil, fmt.Errorf("attachment %s is not a verified default-credential connection pool; check the attachment ID before managing it", id)
	}
	addon, err := client.AddOnInfo(ctx, attachment.Addon.ID)
	if err != nil {
		return nil, err
	}
	if addon.AddonService.Name != "heroku-postgresql" {
		return nil, fmt.Errorf("attachment %s belongs to %q, expected heroku-postgresql", id, addon.AddonService.Name)
	}
	return attachment, nil
}

func resourceHerokuxPostgresConnectionPoolingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	attachment, err := getPoolingAttachment(ctx, meta.(*Config).PlatformAPI, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if attachment == nil {
		d.SetId("")
		return nil
	}
	return setPoolingAttachment(d, attachment)
}

func setPoolingAttachment(d *schema.ResourceData, attachment *heroku.AddOnAttachment) diag.Diagnostics {
	postgresRef := attachment.Addon.ID
	if configured := d.Get("postgres_id").(string); configured != "" && configured == attachment.Addon.Name {
		postgresRef = configured
	}
	for key, value := range map[string]string{"postgres_id": postgresRef, "app_id": attachment.App.ID, "name": attachment.Name, "config_var": attachment.Name + "_URL"} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceHerokuxPostgresConnectionPoolingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := meta.(*Config)
	client := config.PlatformAPI

	log.Printf("[DEBUG] Deleting postgres connection pooling %s", d.Id())

	attachment, err := getPoolingAttachment(ctx, client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if attachment == nil {
		d.SetId("")
		return nil
	}
	deleteErr := client.Delete(ctx, nil, "/addon-attachments/"+url.PathEscape(d.Id()))
	if deleteErr != nil && !platformNotFound(deleteErr) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to removing connection pooling (attachment) %s", d.Id()),
			Detail:   deleteErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Deleted postgres connection pooling %s", d.Id())

	d.SetId("")

	return nil
}

func platformNotFound(err error) bool {
	var apiErr heroku.Error
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}
