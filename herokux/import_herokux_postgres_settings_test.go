package herokux

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresSettings_importBasic(t *testing.T) {
	postgresID := testAccConfig.GetAddonIDorSkip(t)
	wait := true
	connections := true
	duration := randInt(-1, 1999)
	statement := "all"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresSettings_basic(postgresID, wait, connections, duration, statement),
			},
			{
				ResourceName:      "herokux_postgres_settings.foobar",
				ImportStateId:     postgresID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
