package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPostgresCredential_importBasic(t *testing.T) {
	postgresID := testAccConfig.GetAddonIDorSkip(t)
	name := fmt.Sprintf("tftest-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresCredential_basic(postgresID, name),
			},
			{
				ResourceName:      "herokux_postgres_credential.foobar",
				ImportStateId:     fmt.Sprintf("%s:%s", postgresID, name),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
