package herokux

import (
	"testing"

	helper "github.com/davidji99/terraform-provider-herokux/helper/test"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var testAccConfig *helper.TestConfig

func init() {
	testAccProvider = New()
	testAccProviders = map[string]*schema.Provider{
		"herokux": testAccProvider,
	}
	testAccConfig = helper.NewTestConfig()
}

func TestProvider(t *testing.T) {
	if err := New().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = New()
}

func testAccPreCheck(t *testing.T) {
	testAccConfig.GetOrAbort(t, helper.TestConfigHerokuxAPIKey)
}
