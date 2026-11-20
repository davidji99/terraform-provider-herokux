package herokux

import (
	"context"
	"fmt"
	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api/platform"
	"github.com/davidji99/tfph"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	ReleaseStatusSucceeded = "succeeded"
	ReleaseStatusPending   = "pending"
	ReleaseStatusError     = "error"
	ReleaseStatusUnknown   = "unknown"
)

func resourceHerokuxAppContainerRelease() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxAppContainerReleaseCreate,
		ReadContext:   resourceHerokuxAppContainerReleaseRead,
		UpdateContext: resourceHerokuxAppContainerReleaseUpdate,
		DeleteContext: resourceHerokuxAppContainerReleaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxAppContainerReleaseImport,
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

			"image_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateImageID,
			},

			"last_applied_image_id": {Type: schema.TypeString, Computed: true},

			//"container": {
			//	Type:     schema.TypeSet,
			//	Required: true,
			//	MinItems: 1,
			//	Elem: &schema.Resource{
			//		Schema: map[string]*schema.Schema{
			//			"image_id": {
			//				Type:     schema.TypeString,
			//				Required: true,
			//			},
			//			"process_type": {
			//				Type:     schema.TypeString,
			//				Required: true,
			//			},
			//		},
			//	},
			//},

			//"app_name": {
			//	Type:     schema.TypeString,
			//	Computed: true,
			//},
		},
	}
}

var fullContainerDigest = regexp.MustCompile(`^sha256:[A-Fa-f0-9]{64}$`)

func validateImageID(v interface{}, k string) (ws []string, errors []error) {
	name := v.(string)
	if !fullContainerDigest.MatchString(name) {
		errors = append(errors, fmt.Errorf("invalid image ID"))
	}
	return
}

func resourceHerokuxAppContainerReleaseImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	compositeID, parseErr := tfph.ParseCompositeID(d.Id(), 3, "|")
	if parseErr != nil {
		return nil, parseErr
	}

	appID := compositeID[0]
	imageID := compositeID[1]
	processType := compositeID[2]

	// Set the resource ID to the appID and process type.
	d.SetId(fmt.Sprintf("%s:%s", appID, processType))

	d.Set("app_id", appID)
	d.Set("process_type", processType)
	if !fullContainerDigest.MatchString(imageID) {
		return nil, fmt.Errorf("invalid import image digest")
	}
	if err := d.Set("last_applied_image_id", imageID); err != nil {
		return nil, err
	}
	if ds := resourceHerokuxAppContainerReleaseRead(ctx, d, meta); ds.HasError() {
		return nil, fmt.Errorf("import container release: %s", ds[0].Summary)
	}
	if d.Id() == "" || !strings.EqualFold(d.Get("image_id").(string), imageID) {
		return nil, fmt.Errorf("the imported digest does not match the deployed %s image", processType)
	}
	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxAppContainerReleaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	releaseErr := releaseContainer(ctx, d, meta)
	if releaseErr != nil {
		return releaseErr
	}

	// Set the resource ID to the appID and process type.
	d.SetId(fmt.Sprintf("%s:%s", d.Get("app_id").(string), d.Get("process_type").(string)))

	return resourceHerokuxAppContainerReleaseRead(ctx, d, meta)
}

func resourceHerokuxAppContainerReleaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	app, process, id, _, err := observedContainerImage(ctx, d, meta.(*Config))
	if err != nil {
		return diag.FromErr(err)
	}
	if id == "" {
		d.SetId("")
		return nil
	}
	// Legacy state held only the last deployed or imported image_id.
	if d.Get("last_applied_image_id") == "" {
		legacy := d.Get("image_id").(string)
		if !fullContainerDigest.MatchString(legacy) {
			return diag.Errorf("release state has no verified image digest; verify and reimport the deployment")
		}
		if err := d.Set("last_applied_image_id", legacy); err != nil {
			return diag.FromErr(err)
		}
	}
	for key, value := range map[string]string{"app_id": app, "process_type": process, "image_id": id} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func observedContainerImage(ctx context.Context, d *schema.ResourceData, cfg *Config) (app, process, digest, etag string, err error) {
	ids, err := tfph.ParseCompositeID(d.Id(), 2)
	if err != nil {
		return "", "", "", "", err
	}
	app, process = ids[0], ids[1]
	formation, resp, err := cfg.API.Platform.GetContainerFormation(ctx, app, process)
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		// A missing process is conclusive only while the app is still visible.
		_, err = cfg.PlatformAPI.AppInfo(ctx, app)
		return app, process, "", "", err
	}
	if err != nil {
		return app, process, "", "", err
	}
	digest = formation.GetDockerImage().GetID()
	if formation.GetType() != process || formation.App.ID == nil || *formation.App.ID != app || !fullContainerDigest.MatchString(digest) {
		return app, process, "", "", fmt.Errorf("Heroku did not identify the deployed image for %s/%s with a full digest; verify the formation before proceeding", app, process)
	}
	return app, process, digest, resp.Header().Get("ETag"), nil
}

func resourceHerokuxAppContainerReleaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if ds := releaseContainer(ctx, d, meta); ds.HasError() {
		return ds
	}
	return resourceHerokuxAppContainerReleaseRead(ctx, d, meta)
}

func resourceHerokuxAppContainerReleaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*Config)
	app, process, current, etag, err := observedContainerImage(ctx, d, cfg)
	if err != nil {
		return diag.FromErr(err)
	}
	if current != "" {
		owned := d.Get("last_applied_image_id").(string)
		if owned == "" {
			owned = d.Get("image_id").(string)
		}
		if !strings.EqualFold(current, owned) {
			return diag.Errorf("preserved %s/%s: deployed image %s differs from recorded image %s. Verify the deployment and explicitly reimport it or remove this entry from state", app, process, current, owned)
		}
		if err := cfg.API.Platform.RemoveContainerFormation(ctx, app, process, etag); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}

func releaseContainer(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	config := meta.(*Config)
	clientAPI := config.API
	platformAPI := config.PlatformAPI
	opts := &platform.FormationDockerBatchUpdateOpts{}
	imageOpts := platform.FormationDockerUpdateOpts{}

	appID := getAppID(d)

	if v, ok := d.GetOk("image_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] image_id: %s", vs)
		imageOpts.DockerImageID = &vs
	}

	if v, ok := d.GetOk("process_type"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] process_type: %s", vs)
		imageOpts.Type = vs
	}

	opts.Updates = []platform.FormationDockerUpdateOpts{imageOpts}

	//if v, ok := d.GetOk("container"); ok {
	//	vL := v.(*schema.Set).List()
	//	for _, c := range vL {
	//		cInfo := c.(map[string]interface{})
	//		opts.Updates = append(opts.Updates, platform.FormationDockerUpdateOpts{
	//			Type:          cInfo["type"].(string),
	//			DockerImageID: cInfo["image_id"].(string),
	//		})
	//	}
	//}

	log.Printf("[DEBUG] Releasing container on app %s with %v", appID, imageOpts)

	// There's an inconsistency with this Platform API variant where it relies on the Formation endpoint
	// to release the docker image instead of a dedicated Release endpoint (like how heroku_app_release functions).
	_, _, updateErr := clientAPI.Platform.FormationContainerBatchUpdate(appID, opts)
	if updateErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("error releasing %s container release on app %s", imageOpts.Type, appID),
			Detail:   updateErr.Error(),
		})
		return diags
	}

	log.Printf("[INFO] Begin checking if %s container release on app %s is successful", imageOpts.Type, appID)

	stateConf := &resource.StateChangeConf{
		Pending: []string{ReleaseStatusPending, ReleaseStatusUnknown},
		Target:  []string{ReleaseStatusSucceeded},
		Refresh: containerReleaseStateRefreshFunc(platformAPI, appID, *imageOpts.DockerImageID, imageOpts.Type),
		Timeout: time.Duration(config.AppContainerReleaseVerifyTimeout) * time.Minute,
		Delay:   5 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for %s container release on app %s to succeed", imageOpts.Type, appID)
	}

	if err := d.Set("last_applied_image_id", *imageOpts.DockerImageID); err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Released %s container on app %s", imageOpts.Type, appID)

	return diags
}

func containerReleaseStateRefreshFunc(client *heroku.Service, appID, imageID, processType string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Retrieve list of recent releases.
		releases, listErr := client.ReleaseList(context.Background(), appID,
			&heroku.ListRange{Descending: true, Field: "version", Max: 20})
		if listErr != nil {
			return nil, ReleaseStatusError, listErr
		}

		// Find the release that was generated from the releasing the docker image to the app.
		// We'll have to construct the release description manually for comparison.
		shortImageID := strings.Split(imageID, ":")[1][0:12]
		targetReleaseDescription := fmt.Sprintf("Deployed %s (%s)", processType, shortImageID)

		for _, r := range releases {
			if r.Description == targetReleaseDescription && r.Status == ReleaseStatusSucceeded {
				log.Printf("[DEBUG] app %s's %s process - status: %s | current: %v",
					appID, processType, r.Status, r.Current)
				return r, r.Status, nil
			}

			if r.Description == targetReleaseDescription && r.Status != ReleaseStatusSucceeded {
				log.Printf("[DEBUG] Still waiting for app %s's %s process - status: %s | current: %v",
					appID, processType, r.Status, r.Current)
				return r, r.Status, nil
			}
		}

		log.Printf("[DEBUG] Unable to find app %s's %s process in releases", appID, processType)

		return releases, ReleaseStatusUnknown, nil
	}
}
