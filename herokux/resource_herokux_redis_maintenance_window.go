package herokux

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceHerokuxRedisMaintenanceWindow() *schema.Resource {
	return maintenanceWindowResource("redis_id", "heroku-redis")
}
