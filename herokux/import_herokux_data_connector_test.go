package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccHerokuxDataConnector_importBasic(t *testing.T) {
	postgresID := testAccConfig.GetAddonIDorSkip(t)
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	name := fmt.Sprintf("tftest-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxDataConnector_basic(postgresID, kafkaID, name),
			},
			{
				ResourceName:      "herokux_data_connector.foobar",
				ImportStateIdFunc: testAccHerokuxDataConnectorImportStateIDFunc("herokux_data_connector.foobar"),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccHerokuxDataConnectorImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("[ERROR] Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["source_app_name"], rs.Primary.Attributes["name"]), nil
	}
}
