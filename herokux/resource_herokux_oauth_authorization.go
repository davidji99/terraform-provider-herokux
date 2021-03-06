package herokux

import (
	"context"
	"fmt"
	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	OneWeekInSeconds = 604800
)

func resourceHerokuxOauthAuthorization() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxOauthAuthorizationCreate,
		ReadContext:   resourceHerokuxOauthAuthorizationRead,
		UpdateContext: resourceHerokuxOauthAuthorizationUpdate,
		DeleteContext: resourceHerokuxOauthAuthorizationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxOauthAuthorizationImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"scope": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"global", "read",
						"write", "read-protected", "write-protected", "identity"}, false),
				},
				Description: "Set custom OAuth scopes",
			},

			"auth_api_key_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "Any word character (letter, number, underscore) string " +
					"representing the API key used to create the new authorization",
				ValidateFunc: validateAuthAPIKeyName,
			},

			"time_to_live": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "Set expiration in seconds. No expiration if not set.",
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Set a custom authorization description",
			},

			"access_token": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Computed:    true,
				Description: "The actual access token value",
			},

			"expires_in": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "How long in seconds before the token expires",
			},

			"token_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The token UUID",
			},
		},
	}
}

func validateAuthAPIKeyName(v interface{}, k string) (ws []string, errors []error) {
	name := v.(string)
	if !regexp.MustCompile(`^\w{1,32}$`).MatchString(name) {
		errors = append(errors, fmt.Errorf("auth_api_key_name may only include words, letters, or underscore "+
			"with max length of 32 characters"))
	}

	return
}

func resourceHerokuxOauthAuthorizationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	result := strings.Split(d.Id(), ":")

	var authID, ttl string
	authAPIKeyName := ""
	switch len(result) {
	case 2:
		authID = result[0]
		ttl = result[1]
	case 3:
		authID = result[0]
		ttl = result[1]
		authAPIKeyName = result[2]
	default:
		return nil, fmt.Errorf("resource import ID should either be " +
			"[<TOKEN_ID>:<TIME_TO_LIVE>:<AUTH_API_KEY_NAME>] or [<TOKEN_ID>:<TIME_TO_LIVE>]")
	}

	// Set attributes to state so the READ function will work properly
	ttlInt, convErr := strconv.Atoi(ttl)
	if convErr != nil {
		return nil, convErr
	}
	d.Set("time_to_live", ttlInt)
	d.SetId(authID)

	if len(result) == 3 {
		d.Set("auth_api_key_name", authAPIKeyName)
	}

	readErr := resourceHerokuxOauthAuthorizationRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf(readErr[0].Detail)
	}

	return []*schema.ResourceData{d}, nil
}

func constructPlatformAPIClient(d *schema.ResourceData, meta interface{}) (*heroku.Service, diag.Diagnostics) {
	var diags diag.Diagnostics
	client := meta.(*Config).PlatformAPI

	if v, ok := d.GetOk("auth_api_key_name"); ok {
		// Check if the associated env variable is set representing the API key of the user account the new token
		// will be created in Heroku. If no variable is set, then use the default PlatformAPI client created using
		// the token sourced from the HEROKU_API_KEY env variable.

		vs := v.(string)

		// Construct the env variable name
		enVarName := fmt.Sprintf("HEROKUX_%s_API_KEY", strings.ToUpper(vs))
		log.Printf("[DEBUG] env variable to fetch custom Heroku API key: %v", enVarName)
		apiKey, isFound := os.LookupEnv(enVarName)

		// If env variable is not set, error out.
		if !isFound {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("%s not found in the environment", enVarName),
				Detail:   fmt.Sprintf("Please define %s env variable with an API key authorized to create additional OAuth authorizations.", enVarName),
			})
		}

		// Otherwise, initialize a PlatformAPI client with the new API key.
		client = heroku.NewService(&http.Client{
			Transport: &heroku.Transport{
				Username:  "", // Email is not required
				Password:  strings.TrimSpace(apiKey),
				UserAgent: UserAgent,
				Transport: heroku.RoundTripWithRetryBackoff{},
			},
		})
	} else {
		log.Printf("[DEBUG] auth_api_key_name not set. This resource will use the default env variable HEROKU_API_KEY for further actions.")
	}

	return client, diags
}

func resourceHerokuxOauthAuthorizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, clientDiags := constructPlatformAPIClient(d, meta)
	if clientDiags.HasError() {
		return clientDiags
	}

	opts := heroku.OAuthAuthorizationCreateOpts{}

	if v, ok := d.GetOk("scope"); ok {
		vl := v.(*schema.Set).List()
		scopes := make([]string, 0)
		for _, l := range vl {
			scopes = append(scopes, l.(string))
		}
		log.Printf("[DEBUG] oauth authorization scope is : %v", scopes)
		opts.Scope = scopes
	}

	//if v, ok := d.GetOk("client"); ok {
	//	vs := v.(string)
	//	log.Printf("[DEBUG] oauth authorization client is : %v", vs)
	//	opts.Client = &vs
	//}

	if v, ok := d.GetOk("time_to_live"); ok {
		vs := v.(int)
		log.Printf("[DEBUG] oauth authorization time_to_live is : %v seconds", vs)
		opts.ExpiresIn = &vs
	}

	if v, ok := d.GetOk("description"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] oauth authorization description is : %v", vs)
		opts.Description = &vs
	}

	log.Printf("[DEBUG] Creating new OAuth authorization")

	newAuth, createErr := client.OAuthAuthorizationCreate(context.TODO(), opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create new OAuth authorization",
			Detail:   createErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Created new OAuth authorization")

	d.SetId(newAuth.ID)

	return resourceHerokuxOauthAuthorizationRead(ctx, d, meta)
}

func resourceHerokuxOauthAuthorizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, clientDiags := constructPlatformAPIClient(d, meta)

	if clientDiags.HasError() {
		return clientDiags
	}

	hasCustomTTL := false
	if _, ok := d.GetOk("time_to_live"); ok {
		hasCustomTTL = true
	}

	t, getErr := client.OAuthAuthorizationInfo(context.TODO(), d.Id())
	if getErr != nil {
		// Handle when an existing oauth authorization has expired and is no longer available remotely.
		// In this scenario, remove the resource from state so it can be recreated without a `terraform state rm`.
		if strings.Contains(getErr.Error(), "Couldn't find that OAuth") && hasCustomTTL {
			d.SetId("")
			return diags
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to retrieve info about OAuth authorization %s", d.Id()),
			Detail:   getErr.Error(),
		})
		return diags
	}

	// Validate expires_in is less than what is defined in the configuration
	// as the Platform API does not return a field with the original specified TTL.
	// We just want to make sure the TTL was applied correctly.
	if v, ok := d.GetOk("time_to_live"); ok {
		ttl := v.(int)
		if ttl <= *t.AccessToken.ExpiresIn {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "OAuth authorization time-to-live/expiration duration not set properly",
				Detail: fmt.Sprintf("The current expiration duration [%d] is greater than the specified [%d] "+
					"one in your configuration. This should not be the case.", ttl, *t.AccessToken.ExpiresIn),
			})
			return diags
		}
	} else {
		// If no time_to_live is specified, make sure the expires_in is `null` or `nil`.
		if t.AccessToken.ExpiresIn != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "OAuth authorization time-to-live/expiration duration not set properly",
				Detail: fmt.Sprintf("Your configuration does not specify a time_to_live value "+
					"but the authorization access token has an expiration duration of %d seconds. "+
					"Please check and confirm.", *t.AccessToken.ExpiresIn),
			})
			return diags
		}
	}

	var expiresIn int
	if t.AccessToken.ExpiresIn == nil {
		expiresIn = 0
	} else {
		expiresIn = *t.AccessToken.ExpiresIn

		// Add a warning message to tell users their token will expire in given period
		// This warning doesn't work due to a bug in Terraform core.
		if *t.AccessToken.ExpiresIn <= OneWeekInSeconds {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "[WARNING] Oauth authorization token is expiring soon",
				Detail:   fmt.Sprintf("Token %s is expiring in %d seconds", d.Id(), *t.AccessToken.ExpiresIn),
			})
		}
	}
	d.Set("expires_in", expiresIn)

	d.Set("scope", t.Scope)
	d.Set("access_token", t.AccessToken.Token)
	d.Set("description", t.Description)
	d.Set("token_id", t.AccessToken.ID)

	return diags
}

func resourceHerokuxOauthAuthorizationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	opts := heroku.OAuthAuthorizationUpdateOpts{}

	client, clientDiags := constructPlatformAPIClient(d, meta)
	if clientDiags.HasError() {
		return clientDiags
	}

	if d.HasChange("description") {
		vs := d.Get("description").(string)
		opts.Description = &vs
	}

	log.Printf("[DEBUG] Updating OAuth authorization %s", d.Id())

	_, updateErr := client.OAuthAuthorizationUpdate(context.TODO(), d.Id(), opts)
	if updateErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to update OAuth authorization %s", d.Id()),
			Detail:   updateErr.Error(),
		})

		return diags
	}

	log.Printf("[DEBUG] Updated OAuth authorization %s", d.Id())

	return resourceHerokuxOauthAuthorizationRead(ctx, d, meta)
}

func resourceHerokuxOauthAuthorizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, clientDiags := constructPlatformAPIClient(d, meta)
	if clientDiags.HasError() {
		return clientDiags
	}

	_, deleteErr := client.OAuthAuthorizationDelete(context.TODO(), d.Id())
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to delete token %s", d.Id()),
			Detail:   deleteErr.Error(),
		})

		return diags
	}

	d.SetId("")

	return nil
}
