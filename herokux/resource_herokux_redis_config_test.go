package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"strconv"
	"testing"
)

func TestAccHerokuxRedisConfig_Basic(t *testing.T) {
	redisID := testAccConfig.GetRedisIDorSkip(t)
	memPolicy := getRandomStringFromSlice(redisMaxmemoryPolicies)
	keyspaceEvents := getRandomStringFromSlice(redisKeystoreEvents)
	timeout := randInt(1, 500)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxRedisConfig_basic(redisID, memPolicy, keyspaceEvents, timeout),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_redis_config.foobar", "redis_id", redisID),
					resource.TestCheckResourceAttr(
						"herokux_redis_config.foobar", "maxmemory_policy", memPolicy),
					resource.TestCheckResourceAttr(
						"herokux_redis_config.foobar", "notify_keyspace_events", keyspaceEvents),
					resource.TestCheckResourceAttr(
						"herokux_redis_config.foobar", "timeout", strconv.Itoa(timeout)),
				),
			},
			{
				Config: testAccCheckHerokuxRedisConfig_DisableKeyspace(redisID, memPolicy, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_redis_config.foobar", "redis_id", redisID),
					resource.TestCheckResourceAttr(
						"herokux_redis_config.foobar", "maxmemory_policy", memPolicy),
					resource.TestCheckResourceAttr(
						"herokux_redis_config.foobar", "notify_keyspace_events", "disabled"),
					resource.TestCheckResourceAttr(
						"herokux_redis_config.foobar", "timeout", "0"),
				),
			},
		},
	})
}

func TestAccHerokuxRedisConfig_NoTimeout(t *testing.T) {
	redisID := testAccConfig.GetRedisIDorSkip(t)
	memPolicy := getRandomStringFromSlice(redisMaxmemoryPolicies)
	keyspaceEvents := getRandomStringFromSlice(redisKeystoreEvents)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckHerokuxRedisConfig_NoTimeout(redisID, memPolicy, keyspaceEvents),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"herokux_redis_config.foobar", "redis_id", redisID),
					resource.TestCheckResourceAttr(
						"herokux_redis_config.foobar", "maxmemory_policy", memPolicy),
					resource.TestCheckResourceAttr(
						"herokux_redis_config.foobar", "notify_keyspace_events", keyspaceEvents),
				),
			},
		},
	})
}

func testAccCheckHerokuxRedisConfig_basic(redisID, memPolicy, keyspaceEvents string, timeout int) string {
	return fmt.Sprintf(`
resource "herokux_redis_config" "foobar" {
	redis_id = "%s"
	maxmemory_policy = "%s"
	notify_keyspace_events = "%s"
	timeout = %d
}
`, redisID, memPolicy, keyspaceEvents, timeout)
}

func testAccCheckHerokuxRedisConfig_NoTimeout(redisID, memPolicy, keyspaceEvents string) string {
	return fmt.Sprintf(`
resource "herokux_redis_config" "foobar" {
	redis_id = "%s"
	maxmemory_policy = "%s"
	notify_keyspace_events = "%s"
}
`, redisID, memPolicy, keyspaceEvents)
}

func testAccCheckHerokuxRedisConfig_DisableKeyspace(redisID, memPolicy string, timeout int) string {
	return fmt.Sprintf(`
resource "herokux_redis_config" "foobar" {
	redis_id = "%s"
	maxmemory_policy = "%s"
	notify_keyspace_events = "disabled"
	timeout = %d
}
`, redisID, memPolicy, timeout)
}
