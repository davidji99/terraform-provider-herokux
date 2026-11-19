package herokux

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api/postgres/advanced"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestPostgresPublication(t *testing.T) {
	f := newAdvancedTestResource(t, "herokux_postgres_publication")
	path := "/logical-replication/publications"
	item := path + "/sales"
	target := `{"type":"schemas","schemas":["public"]}`
	args := map[string]any{"postgres_id": advancedTestDB, "name": "sales", "schemas": []any{"public"}}
	f.expect("POST", path, `{"name":"sales","target":`+target+`}`, 201, "")
	f.apply(args)
	require.Equal(t, advancedTestDB+":sales", f.state.ID)

	f.expect("GET", item, "", 200, `{"name":"sales","target":`+target+`,"current_tables":["public.new_table"]}`)
	f.importID(f.state.ID)
	require.True(t, f.plan(args).Empty())
	f.expect("GET", item, "", 200, `{"name":"sales","target":{"type":"tables","tables":["public.other"]}}`)
	f.refresh()
	require.False(t, f.plan(args).Empty())
	f.expect("PUT", item, `{"target":`+target+`}`, 204, "")
	f.apply(args)
	for _, status := range []int{403, 500} {
		f.expect("GET", item, "", status, "")
		_, ds := f.resource.RefreshWithoutUpgrade(t.Context(), f.state, f.meta)
		require.True(t, ds.HasError())
		require.Equal(t, advancedTestDB+":sales", f.state.ID)
	}
	f.expect("DELETE", item, "", 204, "")
	f.destroy()
}

func TestAccHerokuxPostgresPublication(t *testing.T) {
	db := advancedAcceptanceDB(t)
	client := advancedAcceptanceClient(t)
	name := "tftest" + acctest.RandString(10)
	const address = "herokux_postgres_publication.test"
	cfg := fmt.Sprintf(`resource "herokux_postgres_publication" "test" {
		postgres_id = %q
		name = %q
		schemas = ["public"]
	}`, db, name)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: cfg, Check: resource.TestCheckResourceAttr(address, "id", db+":"+name)},
			{ResourceName: address, ImportState: true, ImportStateVerify: true},
		},
		CheckDestroy: func(_ *terraform.State) error {
			_, err := client.GetPublication(context.Background(), db, name)
			if errors.Is(err, advanced.ErrNotFound) {
				return nil
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("publication %s still exists", name)
		},
	})
}
