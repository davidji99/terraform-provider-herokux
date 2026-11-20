package herokux

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/data"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHerokuxPrivatelink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPrivatelinkCreate,
		ReadContext:   resourceHerokuxPrivatelinkRead,
		UpdateContext: resourceHerokuxPrivatelinkUpdate,
		DeleteContext: resourceHerokuxPrivatelinkDelete,

		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			desired := d.Get("allowed_accounts").(*schema.Set)
			observed, _ := d.GetChange("observed_allowed_accounts")
			pending, _ := d.GetChange("pending_account_additions")
			if d.HasChange("allowed_accounts") || !desired.Equal(observed) || pending.(*schema.Set).Len() > 0 {
				if err := d.SetNewComputed("observed_allowed_accounts"); err != nil {
					return err
				}
				return d.SetNewComputed("pending_account_additions")
			}
			return nil
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPrivatelinkImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"addon_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"allowed_accounts": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"observed_allowed_accounts": {
				Type: schema.TypeSet, Computed: true, Elem: &schema.Schema{Type: schema.TypeString},
			},

			"pending_account_additions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"service_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func resourceHerokuxPrivatelinkImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())

	readErr := resourceHerokuxPrivatelinkRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf("unable to import %s", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPrivatelinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	addonID := getAddonID(d)
	opts := &postgres.PrivatelinkRequest{}

	if v, ok := d.GetOk("allowed_accounts"); ok {
		aaRaw := v.(*schema.Set).List()
		allowedAccounts := make([]string, 0)

		for _, aa := range aaRaw {
			allowedAccounts = append(allowedAccounts, aa.(string))
		}

		opts.AllowedAccounts = allowedAccounts
		log.Printf("[DEBUG] allowed accounts are : %v", allowedAccounts)
	}

	log.Printf("[DEBUG] Creating Privatelink on addon %s", addonID)

	pl, _, createErr := client.Postgres.CreatePrivatelink(addonID, opts)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	if pl.GetAddon().GetUUID() == "" {
		return diag.Errorf("Heroku returned no ID after creating the PrivateLink")
	}
	d.SetId(pl.GetAddon().GetUUID())

	log.Printf("[DEBUG] Waiting for privatelink on %s to be provisioned", addonID)

	wait := &resource.StateChangeConf{
		Pending:      []string{data.PrivatelinkStatuses.PROVISIONING.ToString()},
		Target:       []string{data.PrivatelinkStatuses.OPERATIONAL.ToString()},
		Refresh:      privatelinkProvisioning(ctx, client, addonID, d.Get("allowed_accounts").(*schema.Set)),
		Timeout:      time.Duration(config.PrivatelinkCreateVerifyTimeout) * time.Minute,
		PollInterval: StateRefreshPollInterval,
	}
	if _, err := wait.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(err)
	}

	return resourceHerokuxPrivatelinkRead(ctx, d, meta)
}

// Keep the observed set separate: the SDK restores a known, configured TypeSet
// from the plan even when Apply returns an error.
func observePrivatelinkAccounts(ctx context.Context, d *schema.ResourceData, client *api.Client) (*postgres.Privatelink, *schema.Set, error) {
	pl, _, err := client.Postgres.GetPrivatelinkContext(ctx, d.Id())
	if err != nil {
		return nil, nil, err
	}
	if pl == nil || pl.AllowedAccounts == nil || !strings.EqualFold(pl.GetAddon().GetUUID(), d.Id()) {
		return nil, nil, fmt.Errorf("incomplete PrivateLink account list for %s", d.Id())
	}
	accounts := schema.NewSet(d.Get("allowed_accounts").(*schema.Set).F, nil)
	for _, account := range pl.AllowedAccounts {
		id := account.GetAccountID()
		if id == "" || accounts.Contains(id) {
			return nil, nil, fmt.Errorf("missing or duplicate PrivateLink account ID on %s", d.Id())
		}
		accounts.Add(id)
	}
	for key, value := range map[string]any{
		"addon_id": d.Id(), "status": pl.Status.ToString(), "service_name": pl.GetServiceName(),
		"observed_allowed_accounts": accounts,
	} {
		if err := d.Set(key, value); err != nil {
			return nil, nil, err
		}
	}
	return pl, accounts, nil
}

func resourceHerokuxPrivatelinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	pl, accounts, err := observePrivatelinkAccounts(ctx, d, meta.(*Config).API)
	if err != nil {
		return diag.FromErr(err)
	}
	if privatelinkAccountsActive(pl, d.Get("pending_account_additions").(*schema.Set)) {
		if err := d.Set("pending_account_additions", nil); err != nil {
			return diag.FromErr(err)
		}
	}
	return diag.FromErr(d.Set("allowed_accounts", accounts))
}

func resourceHerokuxPrivatelinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*Config)
	oldRaw, newRaw := d.GetChange("allowed_accounts")
	desired := newRaw.(*schema.Set)
	observed, _ := d.GetChange("observed_allowed_accounts")
	pending, _ := d.GetChange("pending_account_additions")
	// Restore the receipts before the first fallible lookup; both were unknown in the plan.
	if err := d.Set("observed_allowed_accounts", observed); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pending_account_additions", pending); err != nil {
		return diag.FromErr(err)
	}
	_, current, err := observePrivatelinkAccounts(ctx, d, cfg.API)
	if err != nil {
		return diag.FromErr(err)
	}
	add, remove := desired.Difference(current), current.Difference(desired)
	previous := observed.(*schema.Set)
	if previous.Len() == 0 {
		previous = oldRaw.(*schema.Set) // State from providers predating observed_allowed_accounts.
	}
	if unplanned := remove.Difference(previous); unplanned.Len() > 0 {
		return diag.Errorf("PrivateLink accounts %v were granted after the saved observation; review a fresh plan before removing them", unplanned.List())
	}
	waitFor := pending.(*schema.Set).Intersection(desired).Union(add).Union(desired.Difference(previous))
	if err = d.Set("pending_account_additions", waitFor); err != nil {
		return diag.FromErr(err)
	}
	if add.Len() > 0 {
		_, _, err = cfg.API.Postgres.AddPrivatelinkAllowedAccounts(d.Id(), &postgres.PrivatelinkRequest{AllowedAccounts: formatSetListToStringList(add.List())})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if waitFor.Len() > 0 {
		err = waitForPrivatelinkAccounts(ctx, d, cfg.API, time.Duration(cfg.PrivatelinkAllowedAccountsAddVerifyTimeout)*time.Minute, func(pl *postgres.Privatelink, _ *schema.Set) bool {
			return privatelinkAccountsActive(pl, waitFor)
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if remove.Len() > 0 {
		if err = ctx.Err(); err != nil {
			return diag.FromErr(err)
		}
		_, _, err = cfg.API.Postgres.RemovePrivatelinkAllowedAccounts(d.Id(), &postgres.PrivatelinkRequest{AllowedAccounts: formatSetListToStringList(remove.List())})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	err = waitForPrivatelinkAccounts(ctx, d, cfg.API, time.Duration(cfg.PrivatelinkAllowedAccountsRemoveVerifyTimeout)*time.Minute, func(pl *postgres.Privatelink, accounts *schema.Set) bool {
		return accounts.Equal(desired) && privatelinkAccountsActive(pl, waitFor)
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(d.Set("pending_account_additions", nil))
}

func waitForPrivatelinkAccounts(ctx context.Context, d *schema.ResourceData, client *api.Client, timeout time.Duration, ready func(*postgres.Privatelink, *schema.Set) bool) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("waiting for PrivateLink accounts on %s: %w", d.Id(), err)
		}
		pl, accounts, err := observePrivatelinkAccounts(ctx, d, client)
		if err != nil {
			return err
		}
		if ready(pl, accounts) {
			return ctx.Err()
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for PrivateLink accounts on %s: %w", d.Id(), ctx.Err())
		case <-time.After(StateRefreshPollInterval):
		}
	}
}

func privatelinkAccountsActive(pl *postgres.Privatelink, requested *schema.Set) bool {
	pending := schema.NewSet(requested.F, requested.List())
	for _, a := range pl.AllowedAccounts {
		if a != nil && a.Status == data.PrivatelinkAllowedAccountStatuses.ACTIVE {
			pending.Remove(a.GetAccountID())
		}
	}
	return pending.Len() == 0
}

func formatSetListToStringList(raw []interface{}) []string {
	formatted := make([]string, 0)
	for _, v := range raw {
		formatted = append(formatted, v.(string))
	}

	return formatted
}

func resourceHerokuxPrivatelinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	log.Printf("[DEBUG] Deleting privatelink on addon %s", d.Id())
	_, _, deleteErr := client.Postgres.DeletePrivatelink(d.Id())
	if deleteErr != nil {
		return diag.FromErr(deleteErr)
	}

	log.Printf("[DEBUG] Waiting for privatelink on %s to be deprovisioned", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending:      []string{data.PrivatelinkStatuses.DEPROVISIONING.ToString()},
		Target:       []string{data.PrivatelinkStatuses.DEPROVISIONED.ToString()},
		Refresh:      PrivatelinkDeleteStateRefreshFunc(client, d.Id()),
		Timeout:      time.Duration(config.PrivatelinkDeleteVerifyTimeout) * time.Minute,
		PollInterval: StateRefreshPollInterval,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for privatelink to be deprovisioned on %s: %s", d.Id(), err.Error())
	}

	d.SetId("")

	return nil
}

func PrivatelinkUpdateStateRefreshFunc(client *api.Client, addonID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pl, _, getErr := client.Postgres.GetPrivatelink(addonID)
		if getErr != nil {
			return nil, data.PrivatelinkAllowedAccountStatuses.UNKNOWN.ToString(), getErr
		}

		if pl == nil {
			return nil, "", fmt.Errorf("empty PrivateLink response for %s", addonID)
		}
		if checkAccountsAreActive(pl) {
			return pl, data.PrivatelinkAllowedAccountStatuses.ACTIVE.ToString(), nil
		}
		return pl, data.PrivatelinkAllowedAccountStatuses.PROVISIONING.ToString(), nil
	}
}

func PrivatelinkDeleteStateRefreshFunc(client *api.Client, addonID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pl, response, getErr := client.Postgres.GetPrivatelink(addonID)
		if getErr != nil {
			if response != nil && response.StatusCode == 404 {
				// 404 means the privatelink was deleted
				return pl, data.PrivatelinkStatuses.DEPROVISIONED.ToString(), nil
			}
			// For all other statuses, return the error.
			return nil, postgres.MTLSCertStatuses.UNKNOWN.ToString(), getErr
		}

		if pl == nil || pl.Status == "" {
			return nil, "", fmt.Errorf("missing PrivateLink status for %s", addonID)
		}
		if pl.Status == data.PrivatelinkStatuses.DEPROVISIONED || pl.Status == data.PrivatelinkStatuses.DEPROVISIONING {
			// When a privatelink is deleted, the GET request still returns a 200 with a privatelink status of 'deprovisioned'.
			// This doesn't mean the privatelink was deleted/deprovisioned fully, so this resource will indicate a status
			// of deprovisioning until the GET request returns a 404.
			log.Printf("[DEBUG] Still waiting for privatelink on %s to be deleted", pl.GetAddon().GetUUID())
			return pl, data.PrivatelinkStatuses.DEPROVISIONING.ToString(), nil
		}

		return nil, "", nil
	}
}

func PrivatelinkCreateStateRefreshFunc(client *api.Client, addonID string) resource.StateRefreshFunc {
	return privatelinkProvisioning(context.Background(), client, addonID, nil)
}

func privatelinkProvisioning(ctx context.Context, client *api.Client, addonID string, desired *schema.Set) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pl, _, err := client.Postgres.GetPrivatelinkContext(ctx, addonID)
		if err != nil {
			return nil, "", err
		}
		if pl == nil || pl.Status == "" {
			return nil, "", fmt.Errorf("missing PrivateLink status for %s", addonID)
		}
		if pl.Status == data.PrivatelinkStatuses.OPERATIONAL && (!checkAccountsAreActive(pl) || (desired != nil && !privatelinkAccountsActive(pl, desired))) {
			return pl, data.PrivatelinkStatuses.PROVISIONING.ToString(), nil
		}
		return pl, pl.Status.ToString(), nil
	}
}

func checkAccountsAreActive(pl *postgres.Privatelink) bool {
	for _, account := range pl.AllowedAccounts {
		if account == nil || account.Status != data.PrivatelinkAllowedAccountStatuses.ACTIVE {
			return false
		}
	}
	return len(pl.AllowedAccounts) != 0
}
