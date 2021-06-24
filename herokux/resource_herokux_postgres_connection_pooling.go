package herokux

import (
	"context"
	"fmt"
	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
	"time"
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
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
	client := meta.(*Config).PlatformAPI

	attachment, listErr := client.AddOnAttachmentInfo(context.TODO(), d.Id())
	if listErr != nil {
		return nil, listErr
	}

	d.SetId(attachment.ID)
	d.Set("postgres_id", attachment.Addon.ID)
	d.Set("app_id", attachment.App.ID)
	d.Set("name", attachment.Name)
	d.Set("config_var", fmt.Sprintf("%s_URL", attachment.Name))

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

	if dbInfo.Values[0] != "Available" {
		return diag.Errorf(fmt.Sprintf("Connection pooling not available on database %s. Status is %s", postgresID, dbInfo.Values[0]))
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

	log.Printf("[DEBUG] Waiting for app %s to be restarted after setting config var", appID)

	releases, listErr := platformAPI.ReleaseList(context.TODO(), appID,
		&heroku.ListRange{Descending: true, Field: "version", Max: 1},
	)
	if listErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to retrieve releases for app %s", appID),
			Detail:   listErr.Error(),
		})
		return diags
	}

	if len(releases) == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to verify app %s restarted", appID),
			Detail:   "could not find relevant release",
		})
		return diags
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"pending"},
		Target:       []string{"succeeded"},
		Refresh:      releaseStateRefreshFunc(platformAPI, appID, releases[0].ID),
		Timeout:      20 * time.Minute,
		PollInterval: 5 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for app %s to be restarted after enabling connection pooling on database %s: %s", appID, postgresID, err.Error())
	}

	log.Printf("[DEBUG] Enabled postgres connection pooling on postgres %s", postgresID)

	d.SetId(attachment.ID)

	return resourceHerokuxPostgresConnectionPoolingRead(ctx, d, meta)
}

func resourceHerokuxPostgresConnectionPoolingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).PlatformAPI

	attachment, getErr := client.AddOnAttachmentInfo(context.TODO(), d.Id())
	if getErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("unable to retrieve attachment %s", d.Id()),
			Detail:   getErr.Error(),
		})
		return diags
	}

	d.Set("postgres_id", attachment.Addon.ID)
	d.Set("app_id", attachment.App.ID)
	d.Set("name", attachment.Name)
	d.Set("config_var", fmt.Sprintf("%s_URL", attachment.Name))

	return nil
}

func resourceHerokuxPostgresConnectionPoolingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := meta.(*Config)
	client := config.PlatformAPI

	log.Printf("[DEBUG] Deleting postgres connection pooling %s", d.Id())

	_, deleteErr := client.AddOnAttachmentDelete(context.TODO(), d.Id())
	if deleteErr != nil {
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

func releaseStateRefreshFunc(client *heroku.Service, appID, releaseID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		release, err := client.ReleaseInfo(context.TODO(), appID, releaseID)

		if err != nil {
			return nil, "", err
		}

		// The type conversion here can be dropped when the vendored version of
		// heroku-go is updated.
		return (*heroku.Release)(release), release.Status, nil
	}
}
