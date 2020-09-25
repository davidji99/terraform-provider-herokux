package herokux

import (
	"fmt"
	helper "github.com/davidji99/terraform-provider-herokux/helper/test"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxDataConnector_Basic(t *testing.T) {
	postgresID := testAccConfig.GetAddonIDorSkip(t)
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	name := fmt.Sprintf("tftest-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxDataConnector_basic(postgresID, kafkaID, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "source_id", postgresID),
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "store_id", kafkaID),
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "name", name),
					helper.TestCheckTypeSetElemAttr("herokux_data_connector.foobar",
						"tables.*", "public.users"),
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "excluded_columns.#", "0"),
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "settings.%", "0"),
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "status", "available"),
				),
			},
		},
	})
}

func TestAccHerokuxDataConnector_Paused(t *testing.T) {
	postgresID := testAccConfig.GetAddonIDorSkip(t)
	kafkaID := testAccConfig.GetKafkaIDorSkip(t)
	name := fmt.Sprintf("tftest-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxDataConnector_basic(postgresID, kafkaID, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "source_id", postgresID),
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "store_id", kafkaID),
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "name", name),
					helper.TestCheckTypeSetElemAttr("herokux_data_connector.foobar",
						"tables.*", "public.users"),
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "excluded_columns.#", "0"),
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "settings.%", "0"),
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "status", "available"),
				),
			},
			{
				Config: testAccCheckHerokuxDataConnector_basicPaused(postgresID, kafkaID, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "status", "paused"),
					resource.TestCheckResourceAttr(
						"herokux_data_connector.foobar", "state", "paused"),
				),
			},
		},
	})
}

func testAccCheckHerokuxDataConnector_basic(postgresID, kafkaID, name string) string {
	return fmt.Sprintf(`
resource "herokux_data_connector" "foobar" {
	source_id = "%s"
	store_id = "%s"
	name = "%s"
	tables = ["public.users"]
}
`, postgresID, kafkaID, name)
}

func testAccCheckHerokuxDataConnector_basicPaused(postgresID, kafkaID, name string) string {
	return fmt.Sprintf(`
resource "herokux_data_connector" "foobar" {
	source_id = "%s"
	store_id = "%s"
	name = "%s"
	state = "paused"
	tables = ["public.users"]
}
`, postgresID, kafkaID, name)
}
