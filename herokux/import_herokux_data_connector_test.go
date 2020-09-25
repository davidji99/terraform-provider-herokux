package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
