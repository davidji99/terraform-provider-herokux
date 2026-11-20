package herokux

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api/postgres/advanced"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestPostgresFollowerPool(t *testing.T) {
	f := newAdvancedTestResource(t, "herokux_postgres_follower_pool")
	const path = "/pools/pool-1"
	const id = advancedTestDB + ":pool-1"
	info := func(level string, count int, status string) {
		f.expect("GET", "/info", "", 200, fmt.Sprintf(`{"pools":[{"name":"leader","id":"primary"},{"name":"analytics","id":"pool-1","status":%q,"expected_level":%q,"expected_count":%d}]}`, status, level, count))
	}
	args := map[string]any{"postgres_id": advancedTestDB, "name": "analytics", "level": "4G-Performance"}
	f.expect("POST", "/pools", `{"name":"analytics","level":"4G-Performance","count":1}`, 201, `{"id":"pool-1"}`)
	f.expect("GET", "/info", "", 200, `{"pools":[]}`)
	info("4G-Performance", 1, "provisioning")
	info("4G-Performance", 1, "available")
	f.apply(args)
	require.Equal(t, id, f.state.ID)
	info("4G-Performance", 1, "available")
	f.importID(id)
	require.True(t, f.plan(args).Empty())

	info("16G-Performance", 2, "available") // out-of-band resize
	f.refresh()
	info("16G-Performance", 2, "available")
	f.expect("PATCH", path, `{"level":"4G-Performance"}`, 202, "")
	info("4G-Performance", 2, "available")
	f.expect("PATCH", path, `{"count":1}`, 500, "")
	state, ds := f.resource.Apply(t.Context(), f.state, f.plan(args), f.meta)
	require.True(t, ds.HasError())
	f.state = state
	require.Equal(t, id, state.ID)
	require.Equal(t, "4G-Performance", state.Attributes["level"])
	require.Equal(t, "2", state.Attributes["instance_count"])

	info("4G-Performance", 2, "available")
	f.expect("PATCH", path, `{"count":1}`, 202, "")
	info("4G-Performance", 1, "modifying")
	info("4G-Performance", 1, "available")
	f.apply(args)
	require.True(t, f.plan(args).Empty())

	info("4G-Performance", 1, "available")
	f.expect("DELETE", path, "", 202, "")
	info("4G-Performance", 1, "available")
	f.expect("GET", "/info", "", 200, `{"pools":[]}`)
	f.destroy()
}

func TestPostgresFollowerPoolProtectsLeader(t *testing.T) {
	f := newAdvancedTestResource(t, "herokux_postgres_follower_pool")
	const id = advancedTestDB + ":primary"
	d := schema.TestResourceDataRaw(t, f.resource.Schema, map[string]any{"postgres_id": advancedTestDB, "name": "analytics", "level": "x", "instance_count": 1})
	d.SetId(id)
	imported, err := f.resource.Importer.StateContext(t.Context(), d, f.meta)
	require.NoError(t, err)
	f.expect("GET", "/info", "", 200, `{"pools":[{"id":"primary","name":"leader"}]}`)
	_, ds := f.resource.RefreshWithoutUpgrade(t.Context(), imported[0].State(), f.meta)
	require.True(t, ds.HasError())

	for _, run := range []schema.UpdateContextFunc{f.resource.UpdateContext, schema.UpdateContextFunc(f.resource.DeleteContext)} {
		f.expect("GET", "/info", "", 200, `{"pools":[{"id":"primary","name":"leader"}]}`)
		require.True(t, run(t.Context(), d, f.meta).HasError())
		require.Equal(t, id, d.Id())
	}
	f.expect("GET", "/info", "", 200, `{"pools":[]}`)
	require.True(t, f.resource.UpdateContext(t.Context(), d, f.meta).HasError())
	for _, status := range []int{403, 500} {
		f.expect("GET", "/info", "", status, "")
		require.True(t, f.resource.ReadContext(t.Context(), d, f.meta).HasError())
		require.Equal(t, id, d.Id())
	}
}

func TestFollowerPoolIncompleteInventory(t *testing.T) {
	f := newAdvancedTestResource(t, "herokux_postgres_follower_pool")
	d := f.resource.Data(nil)
	const id = advancedTestDB + ":pool-1"
	d.SetId(id)
	for _, body := range []string{
		`{}`, `{"pools":null}`, `{"pools":[null]}`,
		`{"pools":[{"id":"pool-1","name":"analytics"},{"name":"unknown"}]}`,
	} {
		f.expect("GET", "/info", "", 200, body)
		require.True(t, f.resource.ReadContext(t.Context(), d, f.meta).HasError())
		f.expect("GET", "/info", "", 200, body)
		require.True(t, f.resource.DeleteContext(t.Context(), d, f.meta).HasError())
		require.Equal(t, id, d.Id())
	}
	f.expect("GET", "/info", "", 200, `{"pools":[]}`)
	require.Empty(t, f.resource.ReadContext(t.Context(), d, f.meta))
	require.Empty(t, d.Id())
}

func TestAccHerokuxPostgresFollowerPool(t *testing.T) {
	db := advancedAcceptanceDB(t)
	level := os.Getenv("HEROKUX_POSTGRES_ADVANCED_FOLLOWER_LEVEL")
	if level == "" {
		t.Skip("set HEROKUX_POSTGRES_ADVANCED_FOLLOWER_LEVEL to provision billable test capacity")
	}
	client := advancedAcceptanceClient(t)
	name := "tftest" + acctest.RandString(10)
	const address = "herokux_postgres_follower_pool.test"
	cfg := fmt.Sprintf(`resource "herokux_postgres_follower_pool" "test" {
		postgres_id = %q
		name = %q
		level = %q
	}`, db, name, level)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders, Steps: []resource.TestStep{
			{Config: cfg, Check: resource.TestCheckResourceAttr(address, "instance_count", "1")},
			{ResourceName: address, ImportState: true, ImportStateVerify: true},
		},
		CheckDestroy: func(s *terraform.State) error {
			rs := s.RootModule().Resources[address]
			if rs == nil || rs.Primary == nil {
				return nil
			}
			_, id, err := advancedChildID(rs.Primary.ID)
			if err != nil {
				return err
			}
			_, err = client.GetFollowerPool(context.Background(), db, id)
			if errors.Is(err, advanced.ErrNotFound) {
				return nil
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("follower pool %s still exists", id)
		},
	})
}
