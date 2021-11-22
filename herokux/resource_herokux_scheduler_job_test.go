package herokux

import (
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/helper/test"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"testing"
)

func TestAccHerokuxSchedulerJob_Basic_EveryTenMin(t *testing.T) {
	testAccConfig.GetRunE2ETestsOrSkip(t)

	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	plan := "scheduler:standard"
	frequency := "every_ten_minutes"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ExternalProviders: externalProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxSchedulerJob_basicWithHerokuResource(appName, orgName, plan, "test",
					"Standard-1X", frequency),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"herokux_scheduler_job.foobar", "app_id"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "command", "test"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "dyno_size", "Standard-1X"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "frequency", "every_ten_minutes"),
				),
			},
		},
	})
}

func TestAccHerokuxSchedulerJob_Basic_EveryHour(t *testing.T) {
	testAccConfig.GetRunE2ETestsOrSkip(t)

	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	plan := "scheduler:standard"
	frequency := "every_hour_at_30"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ExternalProviders: externalProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxSchedulerJob_basicWithHerokuResource(appName, orgName, plan, "test",
					"Standard-1X", frequency),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"herokux_scheduler_job.foobar", "app_id"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "command", "test"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "dyno_size", "Standard-1X"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "frequency", "every_hour_at_30"),
				),
			},
		},
	})
}

func TestAccHerokuxSchedulerJob_Basic_EveryDay(t *testing.T) {
	testAccConfig.GetRunE2ETestsOrSkip(t)

	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	plan := "scheduler:standard"
	frequency := "every_day_at_17:30"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ExternalProviders: externalProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxSchedulerJob_basicWithHerokuResource(appName, orgName, plan, "test", "Standard-1X", frequency),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"herokux_scheduler_job.foobar", "app_id"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "command", "test"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "dyno_size", "Standard-1X"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "frequency", "every_day_at_17:30"),
				),
			},
		},
	})
}

func TestAccHerokuxSchedulerJob_Modification(t *testing.T) {
	testAccConfig.GetRunE2ETestsOrSkip(t)

	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	plan := "scheduler:standard"
	frequency := "every_day_at_4:30"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ExternalProviders: externalProviders(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxSchedulerJob_basicWithHerokuResource(appName, orgName, plan, "test", "Standard-1X", frequency),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"herokux_scheduler_job.foobar", "app_id"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "command", "test"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "dyno_size", "Standard-1X"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "frequency", "every_day_at_4:30"),
				),
			},
			{
				Config: testAccCheckHerokuxSchedulerJob_basicWithHerokuResource(appName, orgName, plan, "test1", "Performance-M", "every_hour_at_40"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"herokux_scheduler_job.foobar", "app_id"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "command", "test1"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "dyno_size", "Performance-M"),
					resource.TestCheckResourceAttr(
						"herokux_scheduler_job.foobar", "frequency", "every_hour_at_40"),
				),
			},
		},
	})
}

func TestAccHerokuxSchedulerJob_Basic_InvalidFrequency(t *testing.T) {
	testAccConfig.GetRunE2ETestsOrSkip(t)

	orgName := testAccConfig.GetAnyOrganizationOrSkip(t)
	appName := fmt.Sprintf("tftest-%s", acctest.RandString(10))
	plan := "scheduler:standard"
	frequency := "every1_second_at_14"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		Providers:         testAccProviders,
		ExternalProviders: externalProviders(),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckHerokuxSchedulerJob_basicWithHerokuResource(appName, orgName, plan, "test", "Standard-1X", frequency),
				ExpectError: regexp.MustCompile("unsupported frequency format"),
			},
		},
	})
}

func testAccCheckHerokuxSchedulerJob_basicWithHerokuResource(appName, orgName, addonPlan, command, dynoSize, frequency string) string {
	return fmt.Sprintf(`
%s

resource "herokux_scheduler_job" "foobar" {
  app_id = heroku_app.foobar.uuid
  command = "%s"
  dyno_size = "%s"
  frequency = "%s"

  depends_on = [heroku_addon.foobar]
}
`, test.HerokuAppAddonBlock(appName, orgName, addonPlan), command, dynoSize, frequency)
}
