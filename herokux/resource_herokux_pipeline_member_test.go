package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxPipelineMember_Basic(t *testing.T) {
	pipelineID := testAccConfig.GetPipelineIDorSkip(t)
	email := testAccConfig.GetUserEmailorSkip(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxPipelineMember_basic(pipelineID, email, "\"view\""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_pipeline_member.foobar", "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(
						"herokux_pipeline_member.foobar", "permissions.#", "1"),
					resource.TestCheckResourceAttr(
						"herokux_pipeline_member.foobar", "email", email),
				),
			},
			{
				Config: testAccCheckHerokuxPipelineMember_basic(pipelineID, email, "\"view\", \"deploy\", \"operate\", \"manage\""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_pipeline_member.foobar", "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(
						"herokux_pipeline_member.foobar", "permissions.#", "4"),
					resource.TestCheckResourceAttr(
						"herokux_pipeline_member.foobar", "email", email),
				),
			},
			{
				Config: testAccCheckHerokuxPipelineMember_basic(pipelineID, email, "\"view\", \"deploy\", \"manage\""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_pipeline_member.foobar", "pipeline_id", pipelineID),
					resource.TestCheckResourceAttr(
						"herokux_pipeline_member.foobar", "permissions.#", "3"),
					resource.TestCheckResourceAttr(
						"herokux_pipeline_member.foobar", "email", email),
				),
			},
		},
	})
}

func testAccCheckHerokuxPipelineMember_basic(pipelineID, email, permissions string) string {
	return fmt.Sprintf(`
resource "herokux_pipeline_member" "foobar" {
	pipeline_id = "%s"
	email = "%s"
	permissions = [%s]
}
`, pipelineID, email, permissions)
}
