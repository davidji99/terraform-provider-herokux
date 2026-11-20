package herokux

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testStructuredMaintenanceWindow(t *testing.T, r *schema.Resource, idAttr, service string) {
	t.Helper()
	const uuid = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	for _, identifier := range []string{uuid, service + "-example-12345"} {
		t.Run(identifier, func(t *testing.T) {
			window := "Mondays 09:30 to 13:30 UTC"
			var writes []map[string]string
			lookups := 0
			getStatus, postStatus := 200, 200
			addonService := service
			dataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, "/data/maintenances/v1/"+uuid+"/window", req.URL.Path)
				assert.Equal(t, "Bearer fixture-token", req.Header.Get("Authorization"))
				w.Header().Set("Content-Type", "application/json")
				switch req.Method {
				case http.MethodGet:
					w.WriteHeader(getStatus)
					_ = json.NewEncoder(w).Encode(map[string]string{"window": window, "previous_window": "ignore me"})
				case http.MethodPost:
					var body map[string]string
					if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
						t.Error(err)
						w.WriteHeader(400)
						return
					}
					writes = append(writes, body)
					if postStatus != 200 {
						w.WriteHeader(postStatus)
						return
					}
					switch body["day_of_week"] {
					case "Monday":
						window = "Mondays 09:30 to 13:30 UTC"
					case "Friday":
						window = "Fridays 23:00 to 03:00 UTC"
					default:
						t.Errorf("unexpected weekday %q", body["day_of_week"])
					}
					_ = json.NewEncoder(w).Encode(map[string]string{"window": window})
				default:
					t.Errorf("unexpected method %s", req.Method)
					w.WriteHeader(400)
				}
			}))
			defer dataServer.Close()
			platformServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				lookups++
				if req.Method != http.MethodGet || req.URL.Path != "/addons/"+identifier {
					t.Errorf("unexpected lookup %s %s", req.Method, req.URL.Path)
					w.WriteHeader(400)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"id":%q,"addon_service":{"name":%q}}`, uuid, addonService)
			}))
			defer platformServer.Close()
			legacy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "unexpected legacy endpoint "+r.URL.Path, 404)
			}))
			defer legacy.Close()
			client, err := api.New(config.PostgresAdvancedBaseURL(dataServer.URL), config.PostgresBaseURL(legacy.URL), config.RedisBaseURL(legacy.URL), config.DataBaseURL(legacy.URL), config.APIToken("fixture-token"))
			require.NoError(t, err)
			platform := heroku.NewService(platformServer.Client())
			platform.URL = platformServer.URL
			meta := &Config{API: client, PlatformAPI: platform}
			args := map[string]any{idAttr: identifier, "window": "Mondays 9:30"}
			var state *terraform.InstanceState
			plan := func() *terraform.InstanceDiff {
				t.Helper()
				diff, planErr := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(args), meta)
				require.NoError(t, planErr)
				return diff
			}
			apply := func() {
				t.Helper()
				updated, ds := r.Apply(t.Context(), state, plan(), meta)
				require.False(t, ds.HasError(), "%v", ds)
				state = updated
				require.Equal(t, identifier, state.ID)
				require.Equal(t, identifier, state.Attributes[idAttr])
			}
			apply()
			require.Equal(t, []map[string]string{{"day_of_week": "Monday", "time_of_day": "09:30"}}, writes)
			require.True(t, plan().Empty(), "single-digit hours must round-trip")
			d := r.Data(nil)
			d.SetId(identifier)
			entries, err := r.Importer.StateContext(t.Context(), d, meta)
			require.NoError(t, err)
			require.Len(t, entries, 1)
			state, ds := r.RefreshWithoutUpgrade(t.Context(), entries[0].State(), meta)
			require.False(t, ds.HasError(), "%v", ds)
			require.Equal(t, identifier, state.ID)
			require.True(t, plan().Empty())

			// The new API returns a dedicated window field, including ranges crossing midnight.
			window = "Sundays 20:30 to Mondays 00:30 UTC"
			state, ds = r.RefreshWithoutUpgrade(t.Context(), state, meta)
			require.False(t, ds.HasError(), "%v", ds)
			require.Equal(t, "Sundays 20:30", state.Attributes["window"])
			args["window"] = "Fridays 23:00"
			require.False(t, plan().RequiresNew())
			apply()
			require.Equal(t, map[string]string{"day_of_week": "Friday", "time_of_day": "23:00"}, writes[1])
			require.True(t, plan().Empty())

			for _, bad := range []string{"", "invalid", "Maintenance not required, window is Fridays 23:00 to 03:00 UTC"} {
				window = bad
				got, diags := r.RefreshWithoutUpgrade(t.Context(), state, meta)
				require.True(t, diags.HasError())
				require.Equal(t, state.ID, got.ID)
				require.Equal(t, state.Attributes["window"], got.Attributes["window"])
			}
			window = "Fridays 23:00 to 03:00 UTC"
			getStatus = http.StatusForbidden
			kept, ds := r.RefreshWithoutUpgrade(t.Context(), state, meta)
			require.True(t, ds.HasError())
			require.Equal(t, state.ID, kept.ID)
			getStatus = 200
			args["window"] = "Mondays 9:30"
			postStatus = http.StatusUnprocessableEntity
			kept, ds = r.Apply(t.Context(), state, plan(), meta)
			require.True(t, ds.HasError())
			require.Equal(t, state.ID, kept.ID)
			require.Equal(t, state.Attributes["window"], kept.Attributes["window"])
			postStatus = 200
			addonService = "another-service"
			beforeWrites := len(writes)
			_, ds = r.Apply(t.Context(), state, plan(), meta)
			require.True(t, ds.HasError())
			require.Len(t, writes, beforeWrites)
			beforeLookups := lookups
			gone, ds := r.Apply(t.Context(), state, &terraform.InstanceDiff{Destroy: true}, meta)
			require.False(t, ds.HasError(), "%v", ds)
			require.Nil(t, gone)
			require.Len(t, writes, beforeWrites)
			require.Equal(t, beforeLookups, lookups)
		})
	}
}
