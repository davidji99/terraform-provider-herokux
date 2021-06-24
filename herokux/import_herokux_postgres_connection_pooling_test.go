package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strings"
	"testing"
)

func TestAccHerokuxPostgresConnectionPooling_importBasic(t *testing.T) {
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	name := strings.ToUpper(fmt.Sprintf("tftest_%s", acctest.RandString(10)))
	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:         testAccProviders,
		ExternalProviders: externalProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPostgresConnectionPooling_basic(orgName, appName, name),
			},
			{
				ResourceName:      "herokux_postgres_connection_pooling.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
