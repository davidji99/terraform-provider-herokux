package herokux

import (
	"context"
	"errors"
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
				if err := diff.SetNewComputed("pending_image_id"); err != nil {
					return err
				}
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
			"pending_image_id":      {Type: schema.TypeString, Computed: true},
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
	formation, _, err := observedContainerFormation(ctx, d.Id(), cfg)
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
	digest := formation.GetDockerImage().GetID()
	pending := d.Get("pending_image_id").(string)
	owned := containerImageOwner(d, digest)
	if owned == "" && pending == "" {
		return diag.Errorf("release state has no verified image digest; verify and reimport the deployment")
	}
	if err := d.Set("last_applied_image_id", owned); err != nil {
		return diag.FromErr(err)
	}
	if pending != "" && strings.EqualFold(digest, pending) {
		if err := d.Set("pending_image_id", ""); err != nil {
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
	if digest == "" {
		return diag.Diagnostics{{Severity: diag.Warning, Summary: "Heroku did not report the deployed image digest", Detail: "Image drift cannot be verified for " + d.Id() + ". Terraform retains the recorded digest."}}
	}
	if expected != "" && !strings.EqualFold(expected, digest) {
		return diag.Errorf("Heroku reports %s for %s after releasing %s", digest, d.Id(), expected)
	}
	return diag.FromErr(d.Set("image_id", digest))
}

// An accepted PATCH can outlive a failed status lookup. Only its full deployed digest confirms ownership.
func containerImageOwner(d *schema.ResourceData, deployed string) string {
	owned := d.Get("last_applied_image_id").(string)
	if pending := d.Get("pending_image_id").(string); pending != "" {
		if strings.EqualFold(pending, deployed) {
			return pending
		}
		return owned
	}
	if owned == "" {
		// Older state only recorded the deployed or imported image_id.
		if legacy := d.Get("image_id").(string); fullContainerDigest.MatchString(legacy) {
			return legacy
		}
	}
	return owned
}

func observedContainerFormation(ctx context.Context, id string, cfg *Config) (*platform.Formation, string, error) {
	ids, err := tfph.ParseCompositeID(id, 2)
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
	formation, etag, err := observedContainerFormation(ctx, d.Id(), cfg)
	if err != nil {
		return diag.FromErr(err)
	}
	if formation != nil {
		current := formation.GetDockerImage().GetID()
		app, process := *formation.App.ID, formation.GetType()
		if current == "" {
			return diag.Errorf("cannot remove %s/%s because Heroku did not report its deployed image digest", app, process)
		}
		owned := containerImageOwner(d, current)
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
		Refresh: containerReleaseStateRefreshFunc(ctx, config, appID, image, process, previous),
		Timeout: time.Duration(config.AppContainerReleaseVerifyTimeout) * time.Minute,
	}

	matchedRelease, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		ds := diag.Errorf("waiting for %s container release on app %s: %s", process, appID, err)
		var terminal *resource.UnexpectedStateError
		failed := errors.As(err, &terminal) && terminal.State == "failed"
		return append(ds, recordContainerReleaseFailure(d, image, failed)...)
	}

	if err := d.Set("pending_image_id", ""); err != nil {
		return diag.FromErr(err)
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
	if ds.HasError() && matchedRelease != true {
		return append(ds, recordContainerReleaseFailure(d, image, false)...)
	}
	d.Partial(false)
	return ds
}

func recordContainerReleaseFailure(d *schema.ResourceData, image string, failed bool) diag.Diagnostics {
	// Partial mode would discard the accepted image. Restore the last observation explicitly.
	previous, _ := d.GetChange("image_id")
	owned, _ := d.GetChange("last_applied_image_id")
	priorPending, _ := d.GetChange("pending_image_id")
	if owned == "" && priorPending == "" {
		owned = previous
	}
	d.Partial(false)
	if previous != "" {
		if err := d.Set("image_id", previous); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("last_applied_image_id", owned); err != nil {
		return diag.FromErr(err)
	}
	pending := image
	if failed {
		pending = priorPending.(string)
		if strings.EqualFold(pending, image) {
			pending = ""
		}
		if owned == "" && pending == "" {
			d.SetId("")
		}
	}
	return diag.FromErr(d.Set("pending_image_id", pending))
}

func containerReleaseStateRefreshFunc(ctx context.Context, cfg *Config, appID, imageID, process string, previous int) resource.StateRefreshFunc {
	desc := fmt.Sprintf("Deployed %s (%s)", process, strings.ToLower(strings.TrimPrefix(imageID, "sha256:")[:12]))
	var releaseID string
	return func() (interface{}, string, error) {
		if releaseID != "" {
			r, err := cfg.PlatformAPI.ReleaseInfo(ctx, appID, releaseID)
			if err != nil {
				return nil, ReleaseStatusError, err
			}
			return true, r.Status, nil
		}
		releases, err := cfg.PlatformAPI.ReleaseList(ctx, appID, &heroku.ListRange{Field: "version", Descending: true, Max: 20})
		if err != nil {
			return nil, ReleaseStatusError, err
		}
		if len(releases) == 0 && previous == 0 {
			return false, ReleaseStatusSucceeded, nil
		}
		unchanged := len(releases) > 0 && releases[0].Version == previous
		historicalFailure := false
		for _, r := range releases {
			if r.Description != desc {
				continue
			}
			if r.Version > previous || r.Status == ReleaseStatusPending || (unchanged && r.Version == previous) {
				releaseID = r.ID
				return true, r.Status, nil
			}
			// A later release may have deployed this image in a multi-process update.
			historicalFailure = r.Status == "failed"
			break
		}
		if unchanged && !historicalFailure {
			// The final formation read verifies the requested image for a no-op PATCH.
			return false, ReleaseStatusSucceeded, nil
		}

		// An unrelated or multi-process release can advance the app version without changing this process.
		formation, _, lookupErr := observedContainerFormation(ctx, appID+":"+process, cfg)
		if lookupErr != nil {
			return nil, ReleaseStatusError, lookupErr
		}
		if strings.EqualFold(formation.GetDockerImage().GetID(), imageID) {
			return true, ReleaseStatusSucceeded, nil
		}
		if unchanged {
			return nil, ReleaseStatusError, fmt.Errorf("cannot verify the requested image %s on %s/%s after an earlier failed release", imageID, appID, process)
		}
		return false, ReleaseStatusUnknown, nil
	}
}
