package herokux

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceHerokuxPostgresMaintenanceWindow() *schema.Resource {
	return maintenanceWindowResource("postgres_id", "heroku-postgresql")
}
