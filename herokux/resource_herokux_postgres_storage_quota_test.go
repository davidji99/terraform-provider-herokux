package herokux

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestPostgresStorageQuota(t *testing.T) {
	f := newAdvancedTestResource(t, "herokux_postgres_storage_quota")
	const path = "/quotas/storage"
	const original = `{"warning_gb":12,"critical_gb":20,"enforcement_action":"restrict"}`
	const desired = `{"warning_gb":null,"critical_gb":30,"enforcement_action":"notify","current_gb":11}`
	args := map[string]any{"postgres_id": advancedTestDB, "critical_gb": 30, "enforcement_action": "notify"}
	f.expect("GET", path, "", 200, original)
	f.expect("PATCH", path, `{"warning_gb":null,"critical_gb":30,"enforcement_action":"notify"}`, 200, desired)
	f.apply(args)
	f.expect("GET", path, "", 200, desired)
	f.importID(advancedTestDB)
	require.True(t, f.plan(args).Empty())
	f.expect("GET", path, "", 200, `{"warning_gb":null,"critical_gb":17,"enforcement_action":"notify"}`)
	f.refresh()
	require.False(t, f.plan(args).Empty())
	for _, status := range []int{403, 500} {
		f.expect("GET", path, "", 200, `{"warning_gb":null,"critical_gb":17,"enforcement_action":"notify"}`)
		f.expect("PATCH", path, `{"critical_gb":30}`, status, "")
		prior, ds := f.resource.Apply(t.Context(), f.state, f.plan(args), f.meta)
		require.True(t, ds.HasError())
		require.Equal(t, "17", prior.Attributes["critical_gb"])
	}
	f.expect("GET", path, "", 200, `{"warning_gb":null,"critical_gb":17,"enforcement_action":"notify"}`)
	f.expect("PATCH", path, `{"critical_gb":30}`, 200, desired)
	f.apply(args)
	f.destroy() // The remote policy survives destroy.
}

func TestAccHerokuxPostgresStorageQuota(t *testing.T) {
	db := advancedAcceptanceDB(t)
	client := advancedAcceptanceClient(t)
	original, err := client.GetStorageQuota(t.Context(), db)
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := client.SetStorageQuota(ctx, db, original); err != nil {
			t.Errorf("restoring original storage quota on %s: %v", db, err)
		}
	})
	const address = "herokux_postgres_storage_quota.test"
	cfg := fmt.Sprintf(`resource "herokux_postgres_storage_quota" "test" {
		postgres_id = %q
		enforcement_action = "notify"
		critical_gb = 500
	}`, db)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders, Steps: []resource.TestStep{
			{Config: cfg, Check: resource.TestCheckResourceAttr(address, "critical_gb", "500")},
			{ResourceName: address, ImportState: true, ImportStateVerify: true},
		},
	})
}
