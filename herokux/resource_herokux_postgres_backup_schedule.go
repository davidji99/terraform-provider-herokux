package herokux

import (
	"context"
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
	"strconv"
	"strings"
)

func resourceHerokuxPostgresBackupSchedule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxPostgresBackupScheduleCreate,
		ReadContext:   resourceHerokuxPostgresBackupScheduleRead,
		UpdateContext: resourceHerokuxPostgresBackupScheduleCreate,
		DeleteContext: resourceHerokuxPostgresBackupScheduleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxPostgresBackupScheduleImport,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"postgres_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"hour": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 23),
			},

			"timezone": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "UTC",
				ValidateFunc: validateBackupScheduleTimezone,
				StateFunc:    func(v interface{}) string { return backupScheduleTimezone(v.(string)) },
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"retain_weeks": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"retain_months": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

// pg:backups:schedule expands these aliases before sending the local backup hour to Heroku.
func backupScheduleTimezone(zone string) string {
	switch strings.ToUpper(zone) {
	case "BST", "GMT":
		return "Europe/London"
	case "CDT", "CST":
		return "America/Chicago"
	case "CEST", "CET":
		return "Europe/Paris"
	case "EDT", "EST":
		return "America/New_York"
	case "MDT", "MST":
		return "America/Boise"
	case "PDT", "PST":
		return "America/Los_Angeles"
	case "Z":
		return "UTC"
	default:
		return zone
	}
}

func validateBackupScheduleTimezone(v interface{}, k string) (ws []string, errors []error) {
	timezone := backupScheduleTimezone(v.(string))
	if !regexp.MustCompile(`^UTC|[a-zA-Z]+/[a-zA-Z_]+$`).MatchString(timezone) {
		errors = append(errors, fmt.Errorf("invalid timezone format: use a full time zone name (Africa/Cairo), UTC, or a Heroku backup CLI alias"))
	}
	return
}

func resourceHerokuxPostgresBackupScheduleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Config).API
	postgresID := d.Id()

	schedule, _, listErr := client.Postgres.ListBackupSchedules(postgresID)
	if listErr != nil {
		return nil, listErr
	}

	// As of December 10th, 2020, it is not possible to have more than one backup schedule for a database,
	// so we will just use the zero index element.
	if len(schedule) != 1 {
		return nil, fmt.Errorf("should only expect one backup schedule for postgres database %s, but got %d", postgresID,
			len(schedule))
	}

	d.SetId(schedule[0].GetID())

	retrieveErr := setBackupScheduleToState(d, postgresID, schedule[0])
	if retrieveErr.HasError() {
		return nil, fmt.Errorf("unable to import backup schedule: %v", retrieveErr[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxPostgresBackupScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	opts := &postgres.BackupScheduleRequest{}

	postgresID := getPostgresID(d)

	if v, ok := d.GetOkExists("hour"); ok {
		vs := v.(int)
		log.Printf("[DEBUG] backup_schedule hour is : %v", vs)
		opts.Hour = strconv.Itoa(vs)
	}

	if v, ok := d.GetOk("timezone"); ok {
		opts.Timezone = backupScheduleTimezone(v.(string))
		log.Printf("[DEBUG] backup_schedule timezone is : %v", opts.Timezone)
	}

	log.Printf("[DEBUG] Creating postgres backup schedule on %s", postgresID)

	bs, _, createErr := client.Postgres.CreateBackupSchedule(postgresID, opts)
	if createErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create backup schedule",
			Detail:   createErr.Error(),
		})
		return diags
	}

	log.Printf("[DEBUG] Created postgres backup schedule on %s", postgresID)

	d.SetId(bs.GetID())

	return readBackupSchedule(d, meta, false)
}

func resourceHerokuxPostgresBackupScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return readBackupSchedule(d, meta, true)
}

func readBackupSchedule(d *schema.ResourceData, meta interface{}, allowMissing bool) diag.Diagnostics {
	postgresID := getPostgresID(d)
	var diags diag.Diagnostics
	client := meta.(*Config).API

	schedules, _, readErr := client.Postgres.ListBackupSchedules(postgresID)
	if readErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to list backup schedules",
			Detail:   readErr.Error(),
		})
		return diags
	}

	for _, s := range schedules {
		if s.GetID() == d.Id() {
			return setBackupScheduleToState(d, postgresID, s)
		}
	}
	if !allowMissing {
		return diag.Errorf("backup schedule %s not found after scheduling", d.Id())
	}
	d.SetId("")
	return nil
}

func setBackupScheduleToState(d *schema.ResourceData, postgresID string, schedule *postgres.BackupSchedule) diag.Diagnostics {
	var diags diag.Diagnostics

	d.Set("postgres_id", postgresID)
	d.Set("timezone", backupScheduleTimezone(schedule.GetTimezone()))
	d.Set("retain_weeks", schedule.GetRetainWeeks())
	d.Set("retain_months", schedule.GetRetainMonths())
	d.Set("name", schedule.GetName())

	hourRaw := schedule.GetHour()
	hourNum, convErr := strconv.Atoi(hourRaw.String())
	if convErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to convert backup schedule hour value during state refresh",
			Detail:   convErr.Error(),
		})
		return diags
	}
	d.Set("hour", hourNum)

	return nil
}

func resourceHerokuxPostgresBackupScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	postgresID := getPostgresID(d)

	_, deleteErr := client.Postgres.DeleteBackupSchedule(postgresID, d.Id())
	if deleteErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete backup schedule",
			Detail:   deleteErr.Error(),
		})
		return diags
	}

	d.SetId("")

	return nil
}
