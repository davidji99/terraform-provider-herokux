package herokux

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
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

func TestRedisConfigImportAndPlan(t *testing.T) {
	const redisID = "57d660e0-3d20-40b7-8d20-e77b95189e5a"
	for _, tc := range []struct{ name, remote, configured string }{
		{"enabled", "AKE", "KEA"}, {"disabled", "", "disabled"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/redis/v0/databases/"+redisID+"/config" {
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
					w.WriteHeader(400)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"timeout":{"value":0},"maxmemory_policy":{"value":"noeviction"},"notify_keyspace_events":{"value":%q}}`, tc.remote)
			}))
			defer srv.Close()
			client, err := api.New(config.RedisBaseURL(srv.URL))
			require.NoError(t, err)
			meta := &Config{API: client}
			r := resourceHerokuxRedisConfig()
			d := r.Data(nil)
			d.SetId(redisID)
			imported, err := r.Importer.StateContext(t.Context(), d, meta)
			require.NoError(t, err)
			require.Len(t, imported, 1)
			state, ds := r.RefreshWithoutUpgrade(t.Context(), imported[0].State(), meta)
			require.False(t, ds.HasError(), "%v", ds)
			require.Equal(t, redisID, state.ID)
			args := map[string]any{"redis_id": redisID, "timeout": 0, "maxmemory_policy": "noeviction", "notify_keyspace_events": tc.configured}
			diff, err := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(args), meta)
			require.NoError(t, err)
			require.True(t, diff.Empty(), "%v", diff)
		})
	}
}
