package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/data"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"time"
)

func resourceHerokuxPrivatelink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPrivatelinkCreate,
		ReadContext:   resourceHerokuxPrivatelinkRead,
		UpdateContext: resourceHerokuxPrivatelinkUpdate,
		DeleteContext: resourceHerokuxPrivatelinkDelete,

		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			if d.HasChange("allowed_accounts") {
				// An interrupted update must be able to record its unverified additions in state.
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

	log.Printf("[DEBUG] Waiting for privatelink on %s to be provisioned", addonID)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{data.PrivatelinkStatuses.PROVISIONING.ToString()},
		Target:       []string{data.PrivatelinkStatuses.OPERATIONAL.ToString()},
		Refresh:      PrivatelinkCreateStateRefreshFunc(client, addonID),
		Timeout:      time.Duration(config.PrivatelinkCreateVerifyTimeout) * time.Minute,
		PollInterval: StateRefreshPollInterval,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for privatelink to be provisioned on %s: %s", addonID, err.Error())
	}

	d.SetId(pl.GetAddon().GetUUID())

	return resourceHerokuxPrivatelinkRead(ctx, d, meta)
}

func resourceHerokuxPrivatelinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	pl, _, readErr := client.Postgres.GetPrivatelink(d.Id())
	if readErr != nil {
		return diag.FromErr(readErr)
	}

	if pl == nil {
		return diag.Errorf("empty PrivateLink response for %s", d.Id())
	}
	pending := d.Get("pending_account_additions").(*schema.Set)
	if privatelinkAccountsActive(pl, pending) {
		if err := d.Set("pending_account_additions", nil); err != nil {
			return diag.FromErr(err)
		}
	}

	// Construct an array to store the allowed accounts account_id
	accountIDs := make([]string, 0)
	for _, aa := range pl.AllowedAccounts {
		accountIDs = append(accountIDs, aa.GetAccountID())
	}

	d.Set("addon_id", pl.GetAddon().GetUUID())
	d.Set("allowed_accounts", accountIDs)
	d.Set("status", pl.Status.ToString())
	d.Set("service_name", pl.GetServiceName())

	return nil
}

func resourceHerokuxPrivatelinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	oldRaw, newRaw := d.GetChange("allowed_accounts")
	old, desired := oldRaw.(*schema.Set), newRaw.(*schema.Set)
	toAdd := desired.Difference(old)
	toRemove := old.Difference(desired)
	// After a failed swap, refresh may already list the new accounts while they are still provisioning.
	pending, _ := d.GetChange("pending_account_additions")
	waitFor := pending.(*schema.Set).Intersection(desired).Union(toAdd)

	if toRemove.Len() > 0 && waitFor.Len() > 0 {
		if err := d.Set("pending_account_additions", waitFor); err != nil {
			return diag.FromErr(err)
		}
	}

	if toAdd.Len() > 0 {
		_, _, err := client.Postgres.AddPrivatelinkAllowedAccounts(d.Id(), &postgres.PrivatelinkRequest{AllowedAccounts: formatSetListToStringList(toAdd.List())})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if waitFor.Len() > 0 {
		stateConf := resource.StateChangeConf{
			Pending:      []string{data.PrivatelinkAllowedAccountStatuses.PROVISIONING.ToString()},
			Target:       []string{data.PrivatelinkAllowedAccountStatuses.ACTIVE.ToString()},
			Refresh:      privatelinkUpdateStateRefreshFunc(client, d.Id(), waitFor),
			Timeout:      time.Duration(config.PrivatelinkAllowedAccountsAddVerifyTimeout) * time.Minute,
			PollInterval: StateRefreshPollInterval,
		}
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("error waiting for account ids to become active on %s: %s", d.Id(), err)
		}
	}
	if toRemove.Len() > 0 {
		if err := ctx.Err(); err != nil {
			return diag.FromErr(err)
		}
		_, _, err := client.Postgres.RemovePrivatelinkAllowedAccounts(d.Id(), &postgres.PrivatelinkRequest{AllowedAccounts: formatSetListToStringList(toRemove.List())})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("pending_account_additions", nil); err != nil {
		return diag.FromErr(err)
	}
	return resourceHerokuxPrivatelinkRead(ctx, d, meta)
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
	return privatelinkUpdateStateRefreshFunc(client, addonID, nil)
}

func privatelinkUpdateStateRefreshFunc(client *api.Client, addonID string, required *schema.Set) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pl, _, getErr := client.Postgres.GetPrivatelink(addonID)
		if getErr != nil {
			return nil, data.PrivatelinkAllowedAccountStatuses.UNKNOWN.ToString(), getErr
		}

		if pl == nil {
			return nil, "", fmt.Errorf("empty PrivateLink response for %s", addonID)
		}
		if required != nil {
			if privatelinkAccountsActive(pl, required) {
				return pl, data.PrivatelinkAllowedAccountStatuses.ACTIVE.ToString(), nil
			}
		} else if checkAccountsAreActive(pl) {
			return pl, data.PrivatelinkAllowedAccountStatuses.ACTIVE.ToString(), nil
		}
		return pl, data.PrivatelinkAllowedAccountStatuses.PROVISIONING.ToString(), nil
	}
}

func PrivatelinkDeleteStateRefreshFunc(client *api.Client, addonID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pl, response, getErr := client.Postgres.GetPrivatelink(addonID)
		if getErr != nil {
			if response.StatusCode == 404 {
				// 404 means the privatelink was deleted
				return pl, data.PrivatelinkStatuses.DEPROVISIONED.ToString(), nil
			}
			// For all other statuses, return the error.
			return nil, postgres.MTLSCertStatuses.UNKNOWN.ToString(), getErr
		}

		if pl.Status.ToString() == data.PrivatelinkStatuses.DEPROVISIONED.ToString() && response.StatusCode == 200 {
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
	return func() (interface{}, string, error) {
		// Check the status of both the privatelink and all allowed accounts.
		pl, _, getErr := client.Postgres.GetPrivatelink(addonID)
		if getErr != nil {
			return nil, data.PrivatelinkStatuses.UNKNOWN.ToString(), getErr
		}

		if pl.Status.ToString() == data.PrivatelinkStatuses.PROVISIONING.ToString() {
			log.Printf("[DEBUG] Still waiting for privatelink on %s to be provisioned", addonID)
			return pl, pl.Status.ToString(), nil
		}

		if !checkAccountsAreActive(pl) {
			log.Printf("[DEBUG] Still waiting for privatelink allowed accounts on %s to become active", addonID)
			return pl, data.PrivatelinkAllowedAccountStatuses.PROVISIONING.ToString(), nil
		}

		return pl, data.PrivatelinkStatuses.OPERATIONAL.ToString(), nil
	}
}

func checkAccountsAreActive(pl *postgres.Privatelink) bool {
	// Loop through all the allowed accounts and gather their statuses.
	// Store statuses in an array and check if the array has 'Provisioning' in it.
	aaStatuses := make([]string, 0)

	for _, aa := range pl.AllowedAccounts {
		if aa.Status.ToString() == data.PrivatelinkAllowedAccountStatuses.PROVISIONING.ToString() {
			aaStatuses = append(aaStatuses, aa.Status.ToString())
		}
	}

	return !stringArrayContains(aaStatuses, data.PrivatelinkAllowedAccountStatuses.PROVISIONING.ToString())
}
