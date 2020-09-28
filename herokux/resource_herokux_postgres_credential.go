package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
	"strings"
	"time"
)

const (
	ValidCredentialNameRegex = `^[a-zA-Z0-9_-]{1,50}$`
)

func resourceHerokuxPostgresCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresCredentialCreate,
		ReadContext:   resourceHerokuxPostgresCredentialRead,
		DeleteContext: resourceHerokuxPostgresCredentialDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresCredentialImport,
		},

		Schema: map[string]*schema.Schema{
			"postgres_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCredentialName,
			},

			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"database": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"host": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"port": {
				Type:      schema.TypeInt,
				Computed:  true,
				Sensitive: true,
			},

			"secrets": {
				Type:      schema.TypeList,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:      schema.TypeString,
							Sensitive: true,
							Computed:  true,
						},

						"password": {
							Type:      schema.TypeString,
							Sensitive: true,
							Computed:  true,
						},

						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func validateCredentialName(v interface{}, k string) (ws []string, errors []error) {
	name := v.(string)
	if !regexp.MustCompile(ValidCredentialNameRegex).MatchString(name) {
		errors = append(errors, fmt.Errorf("invalid name: name is restricted to alphanumeric characters(- and _ are also supported) and up to 50 characters"))
	}

	return
}

func resourceHerokuxPostgresCredentialImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Check to make sure the credential name is not 'default' as the 'default' credential cannot be destroyed.
	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return nil, parseErr
	}

	credName := result[1]

	if strings.ToLower(credName) == "default" {
		return nil, fmt.Errorf("cannot import the 'default' credential")
	}

	d.SetId(d.Id())

	readErr := resourceHerokuxPostgresCredentialRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf("unable to import this credential")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPostgresCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	var postgresID, name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
		log.Printf("[DEBUG] credential name is : %v", name)
	}

	if v, ok := d.GetOk("postgres_id"); ok {
		postgresID = v.(string)
		log.Printf("[DEBUG] credential postgres_id is : %v", postgresID)
	}

	log.Printf("[DEBUG] Creating postgres credential %s", name)

	_, _, createErr := client.Postgres.CreateCredential(postgresID, name)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Printf("[DEBUG] Waiting for postgres credential %s to be active", name)

	stateConf := &resource.StateChangeConf{
		Pending: []string{postgres.CredentialStates.WAITFORPROVISIONING.ToString(),
			postgres.CredentialStates.PROVISIONING.ToString()},
		Target:       []string{postgres.CredentialStates.ACTIVE.ToString()},
		Refresh:      postgresCredentialCreationStateRefreshFunc(client, postgresID, name),
		Timeout:      time.Duration(config.PostgresCredentialCreateTimeout) * time.Minute,
		PollInterval: 10 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for postgres credential %s to be active on %s: %s", name, postgresID, err.Error())
	}

	// Set the ID to be a composite of the postgres ID and the credential name.
	d.SetId(fmt.Sprintf("%s:%s", postgresID, name))

	return resourceHerokuxPostgresCredentialRead(ctx, d, meta)
}

func resourceHerokuxPostgresCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	postgresID := result[0]
	credName := result[1]

	cred, _, getErr := client.Postgres.GetCredential(postgresID, credName)
	if getErr != nil {
		return diag.FromErr(getErr)
	}

	d.Set("postgres_id", postgresID)
	d.Set("name", cred.GetName())
	d.Set("state", cred.State.ToString())
	d.Set("database", cred.GetDatabase())
	d.Set("host", cred.GetHost())
	d.Set("port", cred.GetPort())
	d.Set("uuid", cred.GetID())

	// Construct secrets attribute
	secrets := make([]interface{}, 0)
	for _, s := range cred.Credentials {
		c := make(map[string]interface{})
		c["username"] = s.GetUser()
		c["password"] = s.GetPassword()
		c["state"] = s.GetState()

		secrets = append(secrets, c)
	}

	d.Set("secrets", secrets)

	return nil
}

func resourceHerokuxPostgresCredentialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	postgresID := result[0]
	credName := result[1]

	log.Printf("[DEBUG] Deleting postgres credential %s", credName)

	_, _, deleteErr := client.Postgres.DeleteCredential(postgresID, credName)
	if deleteErr != nil {
		return diag.FromErr(deleteErr)
	}

	log.Printf("[DEBUG] Waiting for postgres credential %s to be deleted", credName)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{postgres.CredentialStates.REVOKING.ToString()},
		Target:       []string{postgres.CredentialStates.DELETED.ToString()},
		Refresh:      postgresCredentialDeletionStateRefreshFunc(client, postgresID, credName),
		Timeout:      time.Duration(config.PostgresCredentialDeleteTimeout) * time.Minute,
		PollInterval: 10 * time.Second,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for postgres credential %s to be deleted on %s: %s", credName, postgresID, err.Error())
	}

	return nil
}

func postgresCredentialCreationStateRefreshFunc(client *api.Client, postgresID, credName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cred, _, getErr := client.Postgres.GetCredential(postgresID, credName)
		if getErr != nil {
			return nil, postgres.CredentialStates.UNKNOWN.ToString(), getErr
		}

		if cred.State == postgres.CredentialStates.WAITFORPROVISIONING {
			log.Printf("[DEBUG] postgres credential %s not yet active", cred.GetName())
			return cred, cred.State.ToString(), nil
		}

		return cred, cred.State.ToString(), nil
	}
}

func postgresCredentialDeletionStateRefreshFunc(client *api.Client, postgresID, credName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cred, response, getErr := client.Postgres.GetCredential(postgresID, credName)
		if getErr != nil {
			if response.StatusCode == 404 {
				return postgres.Credential{}, postgres.CredentialStates.DELETED.ToString(), nil
			}
			return nil, postgres.CredentialStates.UNKNOWN.ToString(), getErr
		}

		if cred.State == postgres.CredentialStates.REVOKING {
			log.Printf("[DEBUG] postgres credential %s not yet deleted", cred.GetName())
			return cred, cred.State.ToString(), nil
		}

		return cred, postgres.CredentialStates.UNKNOWN.ToString(), fmt.Errorf("credential not properly deleted")
	}
}
