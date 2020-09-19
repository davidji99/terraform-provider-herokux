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
	heroku "github.com/heroku/heroku-go/v5"
	"log"
	"regexp"
	"strings"
	"time"
)

const (
	Leader   = "leader"
	follower = "follower"
)

func resourceHerokuxPostgres() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresCreate,
		ReadContext:   resourceHerokuxPostgresRead,
		UpdateContext: resourceHerokuxPostgresUpdate,
		DeleteContext: resourceHerokuxPostgresDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresImport,
		},

		Schema: map[string]*schema.Schema{
			"database": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 2, // Increase this later on to support multiple followers for a leader.
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"position": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{Leader, follower}, false),
						},

						"app_id": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsUUID,
						},

						"plan": {
							// Value required is the plan itself sans the 'heroku-postgresql:' part.
							Type:      schema.TypeString,
							Required:  true,
							StateFunc: appendPlanName,
						},

						//"name": {
						//	Type:         schema.TypeString,
						//	Optional:     true,
						//	Computed:     true,
						//	ValidateFunc: validateCustomAddonName,
						//},

						//"config": {
						//	Type:     schema.TypeMap,
						//	Optional: true,
						//	ForceNew: true,
						//},

						"config_vars": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
						},

						"id": {
							Type:        schema.TypeString,
							Description: "The addon ID for the database",
							Computed:    true,
						},

						"name": {
							Type:        schema.TypeString,
							Description: "The addon name for the database",
							Computed:    true,
						},

						//"addon_attachment_id": {
						//	Type:     schema.TypeString,
						//	Computed: true,
						//},
					},
				},
			},

			"database_leader_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"database_follower_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"database_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func appendPlanName(p interface{}) string {
	if p == nil || p == (*string)(nil) {
		return ""
	}
	return fmt.Sprintf("heroku-postgresql:%s", p.(string))
}

func resourceHerokuxPostgresImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return nil, nil
}

func resourceHerokuxPostgresCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	api := config.API
	platformAPI := config.PlatformAPI

	// Define variables to create new leader & follower datbaases
	leaderOpts := heroku.AddOnCreateOpts{}
	var followerOpts heroku.AddOnCreateOpts

	// Define variables to store addon app IDs.
	var leaderAppID string
	var followerAppID string

	// Track if a follower needs to be created
	createFollower := false

	// Validate to make sure there's one database block set to `Leader`.
	if v, ok := d.GetOk("database"); ok {
		vl := v.(*schema.Set).List()

		// Collect all positions
		dbPositions := make([]string, 0)
		for _, db := range vl {
			dbInfo := db.(map[string]interface{})
			if pRaw, ok := dbInfo["position"]; ok {
				dbPositions = append(dbPositions, pRaw.(string))
			}
		}

		log.Printf("[DEBUG] List of database positions : %v", dbPositions)

		// Check if Leader position is present. If not, error out.
		if !stringArrayContains(dbPositions, Leader) {
			return diag.Errorf("did not specify a database with position of '%s' even if you're only creating one database", Leader)
		}

		// Collect information regarding the Leader database.
		// There will always be a Leader database so no need to do a nil check.
		leaderInfo := getDatabaseInfo(vl, Leader)
		if appIdRaw, ok := leaderInfo["app_id"]; ok {
			leaderAppID = appIdRaw.(string)
			log.Printf("[DEBUG] database leader app_id : %v", leaderAppID)
			leaderOpts.Confirm = &leaderAppID
		}

		if planRaw, ok := leaderInfo["plan"]; ok {
			plan := planRaw.(string)
			log.Printf("[DEBUG] database leader plan : %v", plan)
			leaderOpts.Plan = plan
		}

		log.Printf("[DEBUG] Database leader create opts : %v", leaderOpts)

		// Collect information regarding the follower database. For now, there will only be one follower database.
		followerInfo := getDatabaseInfo(vl, follower)
		createFollower = followerInfo != nil
		if createFollower {
			if appIdRaw, ok := followerInfo["app_id"]; ok {
				followerAppID = appIdRaw.(string)
				log.Printf("[DEBUG] database follower app_id : %v", followerAppID)
				followerOpts.Confirm = &followerAppID
			}

			if planRaw, ok := followerInfo["plan"]; ok {
				plan := planRaw.(string)
				log.Printf("[DEBUG] database follower plan : %v", plan)
				followerOpts.Plan = plan
			}
		} else {
			log.Printf("[DEBUG] No database follower defined. Skipping...")
		}

		log.Printf("[DEBUG] Database follower create opts : %v", followerOpts)
	}

	// Now proceed to create the database leader.
	log.Printf("[DEBUG] Creating database leader...")
	leaderDB, leaderCreateErr := platformAPI.AddOnCreate(context.TODO(), leaderAppID, leaderOpts)
	if leaderCreateErr != nil {
		return diag.FromErr(leaderCreateErr)
	}

	log.Printf("[INFO] Database leader ID: %s", leaderDB.ID)

	// Wait for the database leader to be provisioned
	log.Printf("[INFO] Waiting for database leader ID (%s) to be provisioned", leaderDB.ID)
	leaderStateConf := &resource.StateChangeConf{
		Pending: []string{"provisioning"},
		Target:  []string{"provisioned"},
		Refresh: AddOnStateRefreshFunc(platformAPI, leaderDB.ID),
		Timeout: 20 * time.Minute,
	}

	if _, err := leaderStateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("Error waiting for database leader (%s) to be provisioned: %s", leaderDB.ID, err)
	}

	// Now proceed to create follower database if applicable.
	followerDBAppID := ""
	followerDBID := ""
	if createFollower {
		// First, make sure leader database is ready to receive a follower
		log.Printf("[INFO] Waiting for database leader ID (%s) to be able to receive followers", leaderDB.ID)
		followStateConf := &resource.StateChangeConf{
			Pending: []string{"Unavailable", "Temporarily Unavailable"},
			Target:  []string{"Available"},
			Refresh: FollowStateRefreshFunc(api, leaderDB.ID),
			Timeout: 20 * time.Minute,
		}

		if _, err := followStateConf.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for database leader (%s) to be ready for followers: %s", leaderDB.ID, err)
		}

		// Second, set the followerOpts.Config `follow` attribute to follow the leader database
		followerOpts.Config = map[string]string{
			// requires the app NAME instead of UUID
			"follow": fmt.Sprintf("%s::%s", leaderDB.App.Name, "DATABASE_URL"), // FIXME: should this be variable be hardcoded?
		}

		log.Printf("[DEBUG] Creating database follower...")
		followerDB, followerCreateErr := platformAPI.AddOnCreate(context.TODO(), followerAppID, followerOpts)
		if followerCreateErr != nil {
			return diag.FromErr(followerCreateErr)
		}

		log.Printf("[INFO] Database follower ID: %s", followerDB.ID)

		// Wait for the database leader to be provisioned
		log.Printf("[INFO] Waiting for database follower ID (%s) to be provisioned", followerDB.ID)
		followerStateConf := &resource.StateChangeConf{
			Pending: []string{"provisioning"},
			Target:  []string{"provisioned"},
			Refresh: AddOnStateRefreshFunc(platformAPI, followerDB.ID),
			Timeout: 20 * time.Minute,
		}

		if _, err := followerStateConf.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("Error waiting for database follower (%s) to be provisioned: %s", followerDB.ID, err)
		}

		followerDBAppID = followerDB.App.ID
		followerDBID = followerDB.ID
	}

	// If a leader & follower are created, set the resource ID be in the following format:
	// - "<LEADER_APP_ID>|<LEADER_DB_ID>:<FOLLOWER_APP_ID>|<FOLLOWER_DB_ID>"
	if createFollower {
		d.SetId(fmt.Sprintf("%s|%s:%s|%s", leaderDB.App.ID, leaderDB.ID, followerDBAppID, followerDBID))
	} else {
		d.SetId(fmt.Sprintf("%s|%s", leaderDB.App.ID, leaderDB.ID))
	}

	return resourceHerokuxPostgresRead(ctx, d, meta)
}

func resourceHerokuxPostgresRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	//api := config.API
	platformAPI := config.PlatformAPI

	var leaderDatabaseID, followerDatabaseID string
	var dbs []map[string]interface{}

	// Parse resource ID
	resourceIDList := strings.Split(d.Id(), ":")

	// Set leader database info in state
	leaderDatabaseID = strings.Split(resourceIDList[0], "|")[1]
	leaderDB, getLErr := platformAPI.AddOnInfo(context.TODO(), leaderDatabaseID)
	if getLErr != nil {
		return diag.FromErr(getLErr)
	}

	leader := map[string]interface{}{
		"position":    Leader,
		"app_id":      leaderDB.App.ID,
		"plan":        leaderDB.Plan.Name,
		"config_vars": leaderDB.ConfigVars,
		"id":          leaderDB.ID,
		"name":        leaderDB.Name,
	}
	dbs = append(dbs, leader)

	if len(resourceIDList) >= 2 {
		followerDatabaseID = strings.Split(resourceIDList[1], "|")[1]
		followerDB, getFErr := platformAPI.AddOnInfo(context.TODO(), followerDatabaseID)
		if getFErr != nil {
			return diag.FromErr(getFErr)
		}

		follower := map[string]interface{}{
			"position":    follower,
			"app_id":      followerDB.App.ID,
			"plan":        followerDB.Plan.Name,
			"config_vars": followerDB.ConfigVars,
			"id":          followerDB.ID,
			"name":        followerDB.Name,
		}
		dbs = append(dbs, follower)
	}

	// Set database_count in state
	d.Set("database_count", len(resourceIDList))

	// Set database leader/follower ID
	d.Set("database_leader_id", leaderDatabaseID)
	d.Set("database_follower_id", followerDatabaseID)

	// Set database information
	d.Set("database", dbs)

	return nil
}

func resourceHerokuxPostgresUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceHerokuxPostgresDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	platformAPI := config.PlatformAPI

	// Split the resource ID by a colon incase a leader and follower were created prior to deletion.
	resourceIDList := strings.Split(d.Id(), ":")

	// Loop through the resource IDs and delete database(s) in reverse so we delete follower(s) first and then the leader.
	for i := len(resourceIDList) - 1; i >= 0; i-- {
		// Extract the app and db id from compositeID.
		ids := strings.Split(resourceIDList[i], "|")
		appID := ids[0]
		dbID := ids[1]

		log.Printf("[INFO] Deleting database ID (%s) on app ID (%s)", dbID, appID)

		// Destroy the app
		_, deleteErr := platformAPI.AddOnDelete(context.TODO(), appID, dbID)
		if deleteErr != nil {
			return diag.FromErr(deleteErr)
		}
	}

	d.SetId("")

	return nil
}

func getDatabaseInfo(dbList []interface{}, position string) map[string]interface{} {
	for _, db := range dbList {
		dbInfo := db.(map[string]interface{})
		if pRaw, pOK := dbInfo["position"]; pOK {
			if pRaw.(string) == position {
				return dbInfo
			}
		}
	}
	return nil
}

func validateCustomAddonName(v interface{}, k string) (ws []string, errors []error) {
	// Check length
	v1 := validation.StringLenBetween(1, 256)
	_, errs1 := v1(v, k)
	for _, err := range errs1 {
		errors = append(errors, err)
	}

	// Check validity
	valRegex := regexp.MustCompile(`^[a-zA-Z][A-Za-z0-9_-]+$`)
	v2 := validation.StringMatch(valRegex, "Invalid custom addon name: must start with a letter and can only contain lowercase letters, numbers, and dashes")
	_, errs2 := v2(v, k)
	for _, err := range errs2 {
		errors = append(errors, err)
	}

	return ws, errors
}

// AddOnStateRefreshFunc returns a resource.StateRefreshFunc that is used to
// watch an AddOn.
func AddOnStateRefreshFunc(platformAPI *heroku.Service, addOnID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		addon, getErr := platformAPI.AddOnInfo(context.TODO(), addOnID)
		if getErr != nil {
			return nil, "", getErr
		}

		// The type conversion here can be dropped when the vendored version of
		// heroku-go is updated.
		return addon, addon.State, nil
	}
}

// FollowStateRefreshFunc checks if a DB is ready to be followed
func FollowStateRefreshFunc(api *api.Client, dbID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		db, _, getErr := api.Postgres.GetDatabase(dbID)
		if getErr != nil {
			return nil, "", getErr
		}

		followInfo := db.FindInfoByName(postgres.DatabaseInfoNames.FORKFOLLOW.ToString())
		if followInfo == nil {
			return nil, "", fmt.Errorf("could not determine status of %s's follow status", dbID)
		}

		if len(followInfo.Values) == 0 {
			return db, "Unavailable", nil
		}

		return db, followInfo.Values[0].(string), nil
	}
}
