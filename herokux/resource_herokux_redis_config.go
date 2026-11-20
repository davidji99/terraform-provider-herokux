package herokux

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/davidji99/terraform-provider-herokux/api/redis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	redisMaxmemoryPolicies = []string{"noeviction", "allkeys-lru", "volatile-lru", "allkeys-random", "volatile-random",
		"volatile-ttl", "allkeys-lfu", "volatile-lfu"}

	redisKeystoreEvents = []string{
		"K", "E", "g", "$", "l", "s", "h", "z", "t", "x", "e", "m", "A", redis.DisableNotifyKeyspaceEvents}
)

// Redis can reorder flags and abbreviate the standard event classes as A.
func equivalentRedisNotificationFlags(a, b string) bool {
	expand := func(flags string) string {
		if flags == redis.DisableNotifyKeyspaceEvents {
			return ""
		}
		return strings.ReplaceAll(flags, "A", "g$lshztxe")
	}
	a, b = expand(a), expand(b)
	for _, flag := range "KEg$lshztxem" {
		if strings.ContainsRune(a, flag) != strings.ContainsRune(b, flag) {
			return false
		}
	}
	return true
}

func resourceHerokuxRedisConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxRedisConfigUpdate,
		ReadContext:   resourceHerokuxRedisConfigRead,
		UpdateContext: resourceHerokuxRedisConfigUpdate,
		DeleteContext: resourceHerokuxRedisConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: resourceTimeouts(),

		Schema: map[string]*schema.Schema{
			"redis_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"maxmemory_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(redisMaxmemoryPolicies, false),
			},

			"notify_keyspace_events": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(disabled|[KEg$lshztxemA]*)$`), "use Redis notification flags, an empty string, or disabled"),
				StateFunc: func(v interface{}) string {
					if v == "" {
						return redis.DisableNotifyKeyspaceEvents
					}
					return v.(string)
				},
				DiffSuppressFunc: func(_ string, old, next string, d *schema.ResourceData) bool {
					return d.Id() != "" && equivalentRedisNotificationFlags(old, next)
				},
				DiffSuppressOnRefresh: true,
			},

			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
		},
	}
}

func resourceHerokuxRedisConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API
	redisID := getRedisID(d)
	opts := &redis.ConfigUpdateRequest{}

	if v, ok := d.GetOk("maxmemory_policy"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] maxmemory_policy is : %v", vs)
		opts.MaxmemoryPolicy = vs
	}

	if v, ok := d.GetOkExists("notify_keyspace_events"); ok {
		vs := v.(string)
		if vs == redis.DisableNotifyKeyspaceEvents {
			// Set the opt value to empty string ""
			vs = ""
		}

		log.Printf("[DEBUG] notify_keyspace_events is : %v", vs)
		opts.NotifyKeyspaceEvents = &vs
	}

	if v, ok := d.GetOkExists("timeout"); ok {
		vs := v.(int)
		log.Printf("[DEBUG] timeout is : %v", vs)
		opts.Timeout = &vs
	}

	log.Printf("[DEBUG] Updating redis configuration(s) for %s", redisID)
	_, _, updateErr := client.Redis.UpdateConfig(redisID, opts)
	if updateErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to update new redis configurations",
			Detail:   updateErr.Error(),
		})

		return diags
	}

	// Set the ID to the redis ID (UUID)
	d.SetId(redisID)

	log.Printf("[DEBUG] Updated redis configuration(s) for %s", redisID)

	return resourceHerokuxRedisConfigRead(ctx, d, meta)
}

func resourceHerokuxRedisConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Config).API

	c, _, getErr := client.Redis.GetConfig(d.Id())
	if getErr != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to retrieve redis configurations",
			Detail:   getErr.Error(),
		})

		return diags
	}

	d.Set("redis_id", d.Id())
	d.Set("maxmemory_policy", c.GetMaxmemoryPolicy().GetValue())
	d.Set("timeout", c.GetTimeout().GetValue())

	nke := redis.DisableNotifyKeyspaceEvents
	if c.GetNotifyKeyspaceEvents().GetValue() != "" {
		nke = c.GetNotifyKeyspaceEvents().GetValue()
	}

	d.Set("notify_keyspace_events", nke)

	return diags
}

func resourceHerokuxRedisConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Not possible to delete a redis config. Existing resource will only be removed from state.")

	d.SetId("")
	return nil
}
