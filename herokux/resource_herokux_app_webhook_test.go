package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxAppWebhook_Basic(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	name := fmt.Sprintf("tftest-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAppWebhook_CustomNameNoSecret(appID, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "level", "notify"),
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "url", "https://example.com/hooks"),
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "event_types.#", "1"),
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "name", name),
					resource.TestCheckResourceAttrSet(
						"herokux_app_webhook.foobar", "app_name"),
					resource.TestCheckResourceAttrSet(
						"herokux_app_webhook.foobar", "signing_secret"),
				),
			},
		},
	})
}

func TestAccHerokuxAppWebhook_BasicWithSecret(t *testing.T) {
	appID := testAccConfig.GetAppIDorSkip(t)
	name := fmt.Sprintf("tftest-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxAppWebhook_CustomNameSecret(appID, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "app_id", appID),
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "level", "notify"),
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "url", "https://example.com/hooks"),
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "event_types.#", "1"),
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "name", name),
					resource.TestCheckResourceAttrSet(
						"herokux_app_webhook.foobar", "app_name"),
					resource.TestCheckResourceAttrSet(
						"herokux_app_webhook.foobar", "secret"),
				),
			},
			{
				Config: testAccCheckHerokuxAppWebhook_CustomNameSecretUpdated(appID, name+"edited"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "level", "sync"),
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "url", "https://example.com/hooks/123"),
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "event_types.#", "2"),
					resource.TestCheckResourceAttr(
						"herokux_app_webhook.foobar", "name", name+"edited"),
				),
			},
		},
	})
}

func testAccCheckHerokuxAppWebhook_CustomNameNoSecret(appID, name string) string {
	return fmt.Sprintf(`
resource "herokux_app_webhook" "foobar" {
	app_id = "%s"
	level = "notify"
	url = "https://example.com/hooks"
	event_types = ["api:addon-attachment"]
	name = "%s"
}
`, appID, name)
}

func testAccCheckHerokuxAppWebhook_CustomNameSecret(appID, name string) string {
	return fmt.Sprintf(`
resource "herokux_app_webhook" "foobar" {
	app_id = "%s"
	level = "notify"
	url = "https://example.com/hooks"
	event_types = ["api:addon-attachment"]
	name = "%s"
	secret = "my_special_secret"
}
`, appID, name)
}

func testAccCheckHerokuxAppWebhook_CustomNameSecretUpdated(appID, name string) string {
	return fmt.Sprintf(`
resource "herokux_app_webhook" "foobar" {
	app_id = "%s"
	level = "sync"
	url = "https://example.com/hooks/123"
	event_types = ["api:domain", "api:sni-endpoint"]
	name = "%s"
	secret = "my_special_secret123"
}
`, appID, name)
}
