package herokux

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validateOAuthTeam = validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*$`), "use a Heroku team name or UUID")

func resourceHerokuxTeamOAuthAuthorization() *schema.Resource {
	return &schema.Resource{
		CreateContext: createTeamOAuthAuthorization, ReadContext: readTeamOAuthAuthorization,
		UpdateContext: updateTeamOAuthAuthorization, DeleteContext: deleteTeamOAuthAuthorization,
		Importer: &schema.ResourceImporter{StateContext: importTeamOAuthAuthorization},
		Schema: map[string]*schema.Schema{
			"team":         {Type: schema.TypeString, Required: true, ForceNew: true, ValidateFunc: validateOAuthTeam},
			"scope":        {Type: schema.TypeSet, Required: true, MinItems: 1, ForceNew: true, Elem: &schema.Schema{Type: schema.TypeString, ValidateFunc: validateOAuthScope}},
			"description":  {Type: schema.TypeString, Optional: true, Computed: true},
			"time_to_live": {Type: schema.TypeInt, Optional: true, ForceNew: true, ValidateFunc: validation.IntBetween(1, 365*24*60*60)},
			"access_token": {Type: schema.TypeString, Computed: true, Sensitive: true},
			"token_id":     {Type: schema.TypeString, Computed: true},
			"expires_in":   {Type: schema.TypeInt, Computed: true},
		},
	}
}

func importTeamOAuthAuthorization(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	team, id, ok := strings.Cut(d.Id(), "/")
	_, teamErrors := validateOAuthTeam(team, "team")
	_, idErrors := validation.IsUUID(id, "authorization ID")
	if !ok || len(teamErrors) != 0 || len(idErrors) != 0 {
		return nil, fmt.Errorf("expected <team>/<authorization-uuid> for team OAuth authorization import")
	}
	if err := d.Set("team", team); err != nil {
		return nil, err
	}
	d.SetId(strings.ToLower(id))
	return []*schema.ResourceData{d}, nil
}

func createTeamOAuthAuthorization(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	team := d.Get("team").(string)
	opts := heroku.OAuthAuthorizationCreateOpts{}
	for _, scope := range d.Get("scope").(*schema.Set).List() {
		opts.Scope = append(opts.Scope, scope.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		opts.Description = heroku.String(v.(string))
	}
	if v, ok := d.GetOk("time_to_live"); ok {
		opts.ExpiresIn = heroku.Int(v.(int))
	}
	auth, err := meta.(*Config).API.Platform.CreateTeamOAuthAuthorization(ctx, team, opts)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strings.ToLower(auth.ID))
	if auth.AccessToken == nil || auth.AccessToken.Token == "" {
		return diag.Errorf("Heroku created team authorization %s without returning its one-time access token", auth.ID)
	}
	if err := d.Set("access_token", auth.AccessToken.Token); err != nil {
		return diag.FromErr(err)
	}
	return setTeamOAuthMetadata(d, auth)
}

func setTeamOAuthMetadata(d *schema.ResourceData, auth *heroku.OAuthAuthorization) diag.Diagnostics {
	if len(auth.Scope) == 0 {
		return diag.Errorf("missing scope for team authorization %s", auth.ID)
	}
	if err := d.Set("scope", auth.Scope); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", auth.Description); err != nil {
		return diag.FromErr(err)
	}
	if token := auth.AccessToken; token != nil {
		if token.ID != "" {
			if err := d.Set("token_id", token.ID); err != nil {
				return diag.FromErr(err)
			}
		}
		if token.ExpiresIn != nil {
			return diag.FromErr(d.Set("expires_in", *token.ExpiresIn))
		}
	}
	return nil
}

func readTeamOAuthAuthorization(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	auth, err := meta.(*Config).API.Platform.FindTeamOAuthAuthorization(ctx, d.Get("team").(string), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if auth == nil {
		d.SetId("")
		return nil
	}
	// Ignore any token text on read. A changed token ID invalidates a secret we captured at create.
	if _, saved := d.GetOk("access_token"); saved && auth.AccessToken != nil && auth.AccessToken.ID != "" {
		if previous, known := d.GetOk("token_id"); known && previous != auth.AccessToken.ID {
			if err := d.Set("access_token", nil); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return setTeamOAuthMetadata(d, auth)
}

func updateTeamOAuthAuthorization(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*Config)
	auth, err := c.API.Platform.FindTeamOAuthAuthorization(ctx, d.Get("team").(string), d.Id())
	if err != nil {
		d.Partial(true)
		return diag.FromErr(err)
	}
	if auth == nil {
		d.SetId("")
		return diag.Errorf("team OAuth authorization disappeared before its description could be updated")
	}
	opts := heroku.OAuthAuthorizationUpdateOpts{Description: heroku.String(d.Get("description").(string))}
	if _, err := c.PlatformAPI.OAuthAuthorizationUpdate(ctx, d.Id(), opts); err != nil {
		d.Partial(true)
		return teamOAuthMutationError("updating", err)
	}
	return readTeamOAuthAuthorization(ctx, d, meta)
}

func deleteTeamOAuthAuthorization(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c := meta.(*Config)
	auth, err := c.API.Platform.FindTeamOAuthAuthorization(ctx, d.Get("team").(string), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if auth != nil {
		// A revoke can return an empty body. Only its HTTP status is needed.
		err = c.PlatformAPI.Delete(ctx, nil, "/oauth/authorizations/"+d.Id())
		var apiErr heroku.Error
		if err != nil && !(errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound) {
			return teamOAuthMutationError("revoking", err)
		}
	}
	d.SetId("")
	return nil
}

func teamOAuthMutationError(action string, err error) diag.Diagnostics {
	var apiErr heroku.Error
	if errors.As(err, &apiErr) {
		return diag.Errorf("error %s team OAuth authorization: HTTP %d", action, apiErr.StatusCode)
	}
	return diag.Errorf("error %s team OAuth authorization: Platform API request failed", action)
}
