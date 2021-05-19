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
		Timeout:      time.Duration(config.PrivatelinkCreateTimeout) * time.Minute,
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

	// The only updateable attribute for this resource is `allowed_accounts`. However, adding and removing `allowed_accounts`
	// require different request types. Adding accounts requires a PUT but removing accounts requires a PATCH.
	// Therefore, this resource needs to keep track of accounts being added and removed concurrently.
	toAdd := make([]string, 0)
	toRemove := make([]string, 0)

	if d.HasChange("allowed_accounts") {
		o, n := d.GetChange("allowed_accounts")
		old := o.(*schema.Set).List()
		oldF := formatSetListToStringList(old)
		new := n.(*schema.Set).List()
		newF := formatSetListToStringList(new)

		// Determine the accounts to be removed
		for _, id := range oldF {
			if !stringArrayContains(newF, id) {
				toRemove = append(toRemove, id)
			}
		}

		// Determine the accounts to be added
		for _, id := range newF {
			if !stringArrayContains(oldF, id) {
				toAdd = append(toAdd, id)
			}
		}
	}

	log.Printf("[DEBUG] List of account IDs to be removed %v", toRemove)
	log.Printf("[DEBUG] List of account IDs to be added %v", toAdd)

	if len(toRemove) > 0 {
		log.Printf("[DEBUG] Removing the following account ids %v from %s", toRemove, d.Id())

		_, _, removeErr := client.Postgres.RemovePrivatelinkAllowedAccounts(d.Id(),
			&postgres.PrivatelinkRequest{AllowedAccounts: toRemove})

		if removeErr != nil {
			return diag.FromErr(removeErr)
		}
		log.Printf("[DEBUG] Removed the following account ids %v from %s", toRemove, d.Id())
	}

	if len(toAdd) > 0 {
		log.Printf("[DEBUG] Adding the following account ids %v to %s", toAdd, d.Id())

		_, _, addErr := client.Postgres.AddPrivatelinkAllowedAccounts(d.Id(),
			&postgres.PrivatelinkRequest{AllowedAccounts: toAdd})

		if addErr != nil {
			return diag.FromErr(addErr)
		}
		log.Printf("[DEBUG] Added the following account ids %v to %s", toAdd, d.Id())

		// The resource only needs to wait for newly added account ids to become `Active`.
		log.Printf("[DEBUG] Waiting for added account IDS on %s to become active", d.Id())

		stateConf := &resource.StateChangeConf{
			Pending:      []string{data.PrivatelinkAllowedAccountStatuses.PROVISIONING.ToString()},
			Target:       []string{data.PrivatelinkAllowedAccountStatuses.ACTIVE.ToString()},
			Refresh:      PrivatelinkUpdateStateRefreshFunc(client, d.Id()),
			Timeout:      time.Duration(config.PrivatelinkAllowedAccountsAddTimeout) * time.Minute,
			PollInterval: StateRefreshPollInterval,
		}

		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("error waiting for account ids to become active on %s: %s", d.Id(), err.Error())
		}
	}

	return resourceHerokuxPrivatelinkRead(ctx, d, meta)
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
		Timeout:      time.Duration(config.PrivatelinkDeleteTimeout) * time.Minute,
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

		if !checkAccountsAreActive(pl) {
			log.Printf("[DEBUG] Still waiting for privatelink allowed accounts on %s to become active", addonID)
			return pl, data.PrivatelinkAllowedAccountStatuses.PROVISIONING.ToString(), nil
		}

		return pl, data.PrivatelinkAllowedAccountStatuses.ACTIVE.ToString(), nil
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
