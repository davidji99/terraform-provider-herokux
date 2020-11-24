package herokux

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccHerokuxRedisConfig_importBasic(t *testing.T) {
	redisID := testAccConfig.GetRedisIDorSkip(t)
	memPolicy := getRandomStringFromSlice(redisMaxmemoryPolicies)
	keyspaceEvents := getRandomStringFromSlice(redisKeystoreEvents)
	timeout := randInt(1, 500)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxRedisConfig_basic(redisID, memPolicy, keyspaceEvents, timeout),
			},
			{
				ResourceName:      "herokux_redis_config.foobar",
				ImportStateId:     redisID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
