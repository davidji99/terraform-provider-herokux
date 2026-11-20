package herokux

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api/platform"
	"github.com/davidji99/tfph"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	ReleaseStatusSucceeded = "succeeded"
	ReleaseStatusPending   = "pending"
	ReleaseStatusError     = "error"
	ReleaseStatusUnknown   = "unknown"
)

func resourceHerokuxAppContainerRelease() *schema.Resource {
	return &schema.Resource{
		CreateContext: releaseContainer,
		ReadContext:   resourceHerokuxAppContainerReleaseRead,
		UpdateContext: releaseContainer,
		DeleteContext: resourceHerokuxAppContainerReleaseDelete,

		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
			if diff.HasChange("image_id") {
				return diff.SetNewComputed("last_applied_image_id")
			}
			return nil
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxAppContainerReleaseImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     validation.IsUUID,
				DiffSuppressFunc: suppressCaseOnlyDiff,
			},

			"process_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(containerProcessPattern, "invalid process type"),
			},

			"image_id": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validateImageID,
				DiffSuppressFunc: suppressCaseOnlyDiff,
			},

			"last_applied_image_id": {Type: schema.TypeString, Computed: true},
		},
	}
}

var containerProcessPattern = regexp.MustCompile(`^[-\w]{1,128}$`)
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

	if _, errs := validation.IsUUID(appID, "app_id"); len(errs) > 0 || !containerProcessPattern.MatchString(processType) {
		return nil, fmt.Errorf("invalid app UUID or process type in import ID")
	}
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

func resourceHerokuxAppContainerReleaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return readContainerRelease(ctx, d, meta.(*Config), "")
}

func readContainerRelease(ctx context.Context, d *schema.ResourceData, cfg *Config, expected string) diag.Diagnostics {
	formation, _, err := observedContainerFormation(ctx, d, cfg)
	if err != nil {
		return diag.FromErr(err)
	}
	if formation == nil {
		if expected != "" {
			return diag.Errorf("formation %s is missing after release", d.Id())
		}
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
	if !strings.EqualFold(d.Get("app_id").(string), *formation.App.ID) {
		if err := d.Set("app_id", *formation.App.ID); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("process_type", formation.GetType()); err != nil {
		return diag.FromErr(err)
	}
	digest := formation.GetDockerImage().GetID()
	if digest == "" {
		return diag.Diagnostics{{Severity: diag.Warning, Summary: "Heroku did not report the deployed image digest", Detail: "Image drift cannot be verified for " + d.Id() + ". Terraform retains the recorded digest."}}
	}
	if expected != "" && !strings.EqualFold(expected, digest) {
		return diag.Errorf("Heroku reports %s for %s after releasing %s", digest, d.Id(), expected)
	}
	return diag.FromErr(d.Set("image_id", digest))
}

func observedContainerFormation(ctx context.Context, d *schema.ResourceData, cfg *Config) (*platform.Formation, string, error) {
	ids, err := tfph.ParseCompositeID(d.Id(), 2)
	if err != nil {
		return nil, "", err
	}
	app, process := ids[0], ids[1]
	formation, resp, err := cfg.API.Platform.GetContainerFormation(ctx, app, process)
	if resp != nil && resp.StatusCode() == http.StatusNotFound {
		// Heroku can also return 404 when the caller has lost access to the app.
		parent, lookupErr := cfg.PlatformAPI.AppInfo(ctx, app)
		if lookupErr == nil && !strings.EqualFold(parent.ID, app) {
			return nil, "", fmt.Errorf("could not verify parent app %s after formation lookup", app)
		}
		return nil, "", lookupErr
	}
	if err != nil {
		return nil, "", err
	}
	if formation.GetType() != process || formation.App.ID == nil || !strings.EqualFold(*formation.App.ID, app) {
		return nil, "", fmt.Errorf("Heroku returned an unexpected formation for %s/%s", app, process)
	}
	if digest := formation.GetDockerImage().GetID(); digest != "" && !fullContainerDigest.MatchString(digest) {
		return nil, "", fmt.Errorf("Heroku returned an invalid image digest for %s/%s", app, process)
	}
	return formation, resp.Header().Get("ETag"), nil
}

func resourceHerokuxAppContainerReleaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*Config)
	formation, etag, err := observedContainerFormation(ctx, d, cfg)
	if err != nil {
		return diag.FromErr(err)
	}
	if formation != nil {
		current := formation.GetDockerImage().GetID()
		app, process := *formation.App.ID, formation.GetType()
		if current == "" {
			return diag.Errorf("cannot remove %s/%s because Heroku did not report its deployed image digest", app, process)
		}
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
	config := meta.(*Config)
	clientAPI := config.API
	platformAPI := config.PlatformAPI
	appID := getAppID(d)
	image := strings.ToLower(d.Get("image_id").(string))
	d.Partial(d.Id() != "")
	process := d.Get("process_type").(string)
	releases, err := platformAPI.ReleaseList(ctx, appID, &heroku.ListRange{Field: "version", Descending: true, Max: 1})
	if err != nil {
		return diag.FromErr(err)
	}
	var previous int
	if len(releases) != 0 {
		previous = releases[0].Version
	}
	if releaseErr := clientAPI.Platform.ReleaseContainerImage(ctx, appID, process, image); releaseErr != nil {
		return diag.FromErr(releaseErr)
	}
	d.SetId(appID + ":" + process)
	log.Printf("[INFO] Begin checking if %s container release on app %s is successful", process, appID)

	stateConf := &resource.StateChangeConf{
		Pending: []string{ReleaseStatusPending, ReleaseStatusUnknown},
		Target:  []string{ReleaseStatusSucceeded},
		Refresh: containerReleaseStateRefreshFunc(ctx, platformAPI, appID, image, process, previous),
		Timeout: time.Duration(config.AppContainerReleaseVerifyTimeout) * time.Minute,
	}

	matchedRelease, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("waiting for %s container release on app %s: %s", process, appID, err)
	}

	if err := d.Set("last_applied_image_id", image); err != nil {
		return diag.FromErr(err)
	}
	// A successful matching release confirms ownership even if the next GET fails.
	if matchedRelease == true {
		d.Partial(false)
	}
	log.Printf("[DEBUG] Released %s container on app %s", process, appID)

	ds := readContainerRelease(ctx, d, config, image)
	if !ds.HasError() {
		d.Partial(false)
	}
	return ds
}

func containerReleaseStateRefreshFunc(ctx context.Context, client *heroku.Service, appID, imageID, process string, previous int) resource.StateRefreshFunc {
	desc := fmt.Sprintf("Deployed %s (%s)", process, strings.ToLower(strings.TrimPrefix(imageID, "sha256:")[:12]))
	var releaseID string
	return func() (interface{}, string, error) {
		if releaseID != "" {
			r, err := client.ReleaseInfo(ctx, appID, releaseID)
			if err != nil {
				return nil, ReleaseStatusError, err
			}
			return true, r.Status, nil
		}
		releases, err := client.ReleaseList(ctx, appID, &heroku.ListRange{Field: "version", Descending: true, Max: 20})
		if err != nil {
			return nil, ReleaseStatusError, err
		}
		unchanged := len(releases) > 0 && releases[0].Version == previous
		for _, r := range releases {
			if r.Description == desc && (r.Version > previous || unchanged) {
				releaseID = r.ID
				return true, r.Status, nil
			}
		}
		if unchanged || (len(releases) == 0 && previous == 0) {
			return false, ReleaseStatusSucceeded, nil
		}
		return false, ReleaseStatusUnknown, nil
	}
}
