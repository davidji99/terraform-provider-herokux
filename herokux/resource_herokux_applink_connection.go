package herokux

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davidji99/terraform-provider-herokux/api/applink"
	"github.com/davidji99/tfph"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var errAppLinkDisconnected = errors.New("AppLink connection is disconnected; reconnect it before importing or applying an update")

const appLinkLoginURL = "https://login.salesforce.com"

func resourceHerokuxAppLinkConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: createAppLinkConnection,
		ReadContext:   readAppLinkConnection,
		UpdateContext: updateAppLinkConnection,
		DeleteContext: deleteAppLinkConnection,
		Importer:      &schema.ResourceImporter{StateContext: importAppLinkConnection},
		Timeouts:      resourceTimeouts(),
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ any) error {
			if d.Id() != "" {
				return nil
			}
			for _, k := range []string{"client_id", "username", "jwt_private_key_wo"} {
				v, ds := d.GetRawConfigAt(cty.GetAttrPath(k))
				if ds.HasError() {
					return fmt.Errorf("could not inspect creation argument %s", k)
				}
				if v.IsNull() {
					return fmt.Errorf("%s is required to create an AppLink connection", k)
				}
				if k == "jwt_private_key_wo" && v.IsKnown() {
					return validateAppLinkPrivateKey(v.AsString())
				}
			}
			return nil
		},
		Schema: map[string]*schema.Schema{
			"app_id":             {Type: schema.TypeString, Required: true, ValidateFunc: validation.IsUUID},
			"addon_id":           {Type: schema.TypeString, Required: true, ForceNew: true, ValidateFunc: validation.IsUUID},
			"name":               {Type: schema.TypeString, Required: true, ForceNew: true, ValidateFunc: validateAppLinkName},
			"client_id":          {Type: schema.TypeString, Optional: true, ValidateFunc: validation.StringIsNotEmpty},
			"username":           {Type: schema.TypeString, Optional: true, ValidateFunc: validation.StringIsNotEmpty},
			"login_url":          {Type: schema.TypeString, Optional: true, ValidateFunc: validation.IsURLWithHTTPS},
			"jwt_private_key_wo": {Type: schema.TypeString, Optional: true, Sensitive: true, WriteOnly: true},
			"connection_id":      {Type: schema.TypeString, Computed: true},
			"org_id":             {Type: schema.TypeString, Computed: true},
			"instance_url":       {Type: schema.TypeString, Computed: true},
			"status":             {Type: schema.TypeString, Computed: true},
		},
	}
}

func validateAppLinkName(v any, k string) ([]string, []error) {
	name := v.(string)
	if strings.TrimSpace(name) == "" || name == "." || name == ".." || strings.ContainsAny(name, ":/") {
		return nil, []error{fmt.Errorf("%s must be a connection name without slashes or colons", k)}
	}
	return nil, nil
}

type appLinkID struct{ app, addon, name string }

func (id appLinkID) String() string { return id.app + ":" + id.addon + ":" + id.name }

func parseAppLinkID(raw string) (appLinkID, error) {
	parts, err := tfph.ParseCompositeID(raw, 3)
	if err != nil {
		return appLinkID{}, err
	}
	for _, v := range parts[:2] {
		if _, errs := validation.IsUUID(v, "app/add-on ID"); len(errs) > 0 {
			return appLinkID{}, errs[0]
		}
	}
	if _, errs := validateAppLinkName(parts[2], "name"); len(errs) > 0 {
		return appLinkID{}, errs[0]
	}
	return appLinkID{parts[0], parts[1], parts[2]}, nil
}

func configuredAppLinkID(d *schema.ResourceData) appLinkID {
	return appLinkID{d.Get("app_id").(string), d.Get("addon_id").(string), d.Get("name").(string)}
}

func appLinkPrivateKey(d *schema.ResourceData) (string, error) {
	v, ds := d.GetRawConfigAt(cty.GetAttrPath("jwt_private_key_wo"))
	if ds.HasError() || !v.IsKnown() || v.IsNull() || !v.Type().Equals(cty.String) {
		return "", errors.New("jwt_private_key_wo is required to create an AppLink connection")
	}
	key := v.AsString()
	return key, validateAppLinkPrivateKey(key)
}

func validateAppLinkPrivateKey(key string) error {
	block, rest := pem.Decode([]byte(key))
	if block != nil && strings.TrimSpace(string(rest)) == "" {
		switch block.Type {
		case "RSA PRIVATE KEY":
			if _, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				return nil
			}
		case "PRIVATE KEY":
			parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if _, ok := parsed.(*rsa.PrivateKey); err == nil && ok {
				return nil
			}
		}
	}
	return errors.New("jwt_private_key_wo must be an unencrypted RSA private key in PEM format")
}

func (id appLinkID) verify(c applink.Connection, expected string) error {
	if c.ID == "" || (expected != "" && c.ID != expected) {
		return fmt.Errorf("connection %q has a different or missing remote ID; verify the connection before importing it", id.name)
	}
	if c.Org.Type != "SalesforceOrg" || c.Org.Name != id.name || (c.AddonID != "" && c.AddonID != id.addon) {
		return fmt.Errorf("AppLink did not return the requested Salesforce org connection %q", id.name)
	}
	switch c.Status {
	case "connected", "connecting", "disconnected", "disconnecting", "failed":
		return nil
	default:
		return fmt.Errorf("AppLink returned an unknown status for connection %q", id.name)
	}
}

func (id appLinkID) save(d *schema.ResourceData, c applink.Connection) error {
	for k, v := range map[string]string{"app_id": id.app, "addon_id": id.addon, "name": id.name, "connection_id": c.ID, "org_id": c.Org.ID, "instance_url": c.Org.InstanceURL, "status": c.Status} {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	d.SetId(id.String())
	return nil
}

func createAppLinkConnection(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	clientID, username := d.Get("client_id").(string), d.Get("username").(string)
	if clientID == "" || username == "" {
		return diag.Errorf("client_id and username are required to create an AppLink connection")
	}
	key, err := appLinkPrivateKey(d)
	if err != nil {
		return diag.FromErr(err)
	}
	login := d.Get("login_url").(string)
	if login == "" {
		login = appLinkLoginURL
	}
	id := configuredAppLinkID(d)
	s, err := meta.(*Config).API.AppLink.Session(ctx, id.app, id.addon)
	if err != nil {
		return diag.FromErr(err)
	}
	existing, lookupErr := s.Get(ctx, id.name)
	switch {
	case lookupErr == nil:
		if err = id.verify(existing, ""); err != nil {
			return diag.FromErr(err)
		}
		if existing.Status != "disconnected" {
			return diag.Errorf("AppLink connection %q already exists; import it to manage it", id.name)
		}
	case !errors.Is(lookupErr, applink.ErrNotFound):
		return diag.FromErr(lookupErr)
	}
	c, err := s.Create(ctx, id.name, clientID, username, login, key)
	if err != nil {
		return diag.FromErr(err)
	}
	if c.ID == "" {
		return diag.Errorf("AppLink accepted the create request without a connection ID; inspect the add-on before retrying")
	}
	d.SetId(id.String())
	if err = d.Set("connection_id", c.ID); err != nil {
		return diag.FromErr(err)
	}
	if err = id.verify(c, ""); err != nil {
		return diag.FromErr(err)
	}
	if err = id.save(d, c); err != nil {
		return diag.FromErr(err)
	}
	switch c.Status {
	case "connected":
		return nil
	case "connecting":
		observed, waitErr := id.wait(ctx, s, c.ID, false)
		ds := diag.FromErr(waitErr)
		if observed.ID != "" {
			ds = append(ds, diag.FromErr(id.save(d, observed))...)
		}
		return ds
	default:
		return diag.Errorf("AppLink connection %q returned status %q after creation", id.name, c.Status)
	}
}

func (id appLinkID) wait(ctx context.Context, s *applink.Session, expected string, deleting bool) (applink.Connection, error) {
	lookup := id.name
	if deleting {
		lookup = expected
	}
	for {
		c, err := s.Get(ctx, lookup)
		switch {
		case errors.Is(err, applink.ErrNotFound):
			if deleting {
				return c, nil
			}
		case err != nil:
			return applink.Connection{}, err
		default:
			if err = id.verify(c, expected); err != nil {
				return applink.Connection{}, err
			}
			if deleting {
				switch c.Status {
				case "disconnected":
					return c, nil
				case "connected", "connecting", "disconnecting":
					// An accepted disconnect can still return the previous status.
				default:
					return c, fmt.Errorf("disconnecting AppLink connection %q failed", id.name)
				}
			} else {
				switch c.Status {
				case "connected":
					return c, nil
				case "connecting":
				default:
					return c, fmt.Errorf("connecting AppLink connection %q failed", id.name)
				}
			}
		}
		select {
		case <-ctx.Done():
			return c, fmt.Errorf("waiting for AppLink connection %q: %w", id.name, ctx.Err())
		case <-time.After(2 * time.Second):
		}
	}
}

func (id appLinkID) refresh(ctx context.Context, client *applink.Client, d *schema.ResourceData, expected string) error {
	s, err := client.Session(ctx, id.app, id.addon)
	if err != nil {
		return err
	}
	c, err := s.Get(ctx, id.name)
	if err != nil {
		return err
	}
	if err = id.verify(c, expected); err != nil {
		return err
	}
	if c.Status == "disconnected" {
		return errAppLinkDisconnected
	}
	return id.save(d, c)
}

func importAppLinkConnection(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	id, err := parseAppLinkID(d.Id())
	if err != nil {
		return nil, err
	}
	if err = id.refresh(ctx, meta.(*Config).API.AppLink, d, ""); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func readAppLinkConnection(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	id, err := parseAppLinkID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	expected := d.Get("connection_id").(string)
	if expected == "" {
		return diag.Errorf("missing remote AppLink connection ID in state; import the connection")
	}
	err = id.refresh(ctx, meta.(*Config).API.AppLink, d, expected)
	if errors.Is(err, applink.ErrNotFound) || errors.Is(err, errAppLinkDisconnected) {
		d.SetId("")
		return nil
	}
	return diag.FromErr(err)
}

func updateAppLinkConnection(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	id := configuredAppLinkID(d)
	expected := d.Get("connection_id").(string)
	if expected == "" {
		return diag.Errorf("missing remote AppLink connection ID in state; import the connection")
	}
	// Verify the new access path before replacing the previous app locator in state.
	if err := id.refresh(ctx, meta.(*Config).API.AppLink, d, expected); err != nil {
		d.Partial(true)
		return diag.FromErr(err)
	}
	return nil
}

func deleteAppLinkConnection(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	id, err := parseAppLinkID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	expected := d.Get("connection_id").(string)
	if expected == "" {
		return diag.Errorf("cannot disconnect an AppLink connection without its recorded remote ID")
	}
	s, err := meta.(*Config).API.AppLink.Session(ctx, id.app, id.addon)
	if err != nil {
		return diag.FromErr(err)
	}
	c, err := s.Get(ctx, id.name)
	if errors.Is(err, applink.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if err = id.verify(c, expected); err != nil {
		return diag.FromErr(err)
	}
	if c.Status != "disconnected" {
		if c.Status != "disconnecting" {
			err = s.Disconnect(ctx, expected)
			if errors.Is(err, applink.ErrNotFound) {
				d.SetId("")
				return nil
			}
			if err != nil {
				return diag.FromErr(err)
			}
		}
		if _, err = id.wait(ctx, s, expected, true); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}
