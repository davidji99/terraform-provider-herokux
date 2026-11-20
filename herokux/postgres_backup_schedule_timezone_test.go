package herokux

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresBackupScheduleTimezoneAliases(t *testing.T) {
	for _, tc := range []struct{ input, zone string }{
		{"PST", "America/Los_Angeles"},
		{"PDT", "America/Los_Angeles"},
		{"pSt", "America/Los_Angeles"},
		{"BST", "Europe/London"},
		{"GMT", "Europe/London"},
		{"CDT", "America/Chicago"},
		{"CST", "America/Chicago"},
		{"CEST", "Europe/Paris"},
		{"CET", "Europe/Paris"},
		{"EDT", "America/New_York"},
		{"EST", "America/New_York"},
		{"MDT", "America/Boise"},
		{"MST", "America/Boise"},
		{"Z", "UTC"},
		{"UTC", "UTC"},
		{"Australia/Perth", "Australia/Perth"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			zone := tc.zone
			writes := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path != "/client/v11/databases/db/transfer-schedules" {
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(404)
					return
				}
				schedule := map[string]any{"uuid": "daily", "name": "DATABASE_URL", "hour": 3, "timezone": zone}
				switch r.Method {
				case http.MethodPost:
					var body map[string]any
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					assert.Equal(t, map[string]any{"hour": "3", "timezone": tc.zone}, body)
					writes++
					_ = json.NewEncoder(w).Encode(schedule)
				case http.MethodGet:
					_ = json.NewEncoder(w).Encode([]any{schedule})
				default:
					t.Errorf("unexpected %s", r.Method)
					w.WriteHeader(405)
				}
			}))
			defer srv.Close()
			client, err := api.New(config.PostgresBaseURL(srv.URL))
			require.NoError(t, err)
			meta := &Config{API: client}
			resource := resourceHerokuxPostgresBackupSchedule()
			cfg := terraform.NewResourceConfigRaw(map[string]any{"postgres_id": "db", "hour": 3, "timezone": tc.input})
			ds := resource.Validate(cfg)
			require.False(t, ds.HasError(), "%v", ds)
			plan, err := resource.Diff(t.Context(), nil, cfg, meta)
			require.NoError(t, err)
			state, ds := resource.Apply(t.Context(), nil, plan, meta)
			require.False(t, ds.HasError(), "%v", ds)
			require.Equal(t, 1, writes)
			state, ds = resource.RefreshWithoutUpgrade(t.Context(), state, meta)
			require.False(t, ds.HasError(), "%v", ds)
			require.Equal(t, tc.zone, state.Attributes["timezone"])
			plan, err = resource.Diff(t.Context(), state, cfg, meta)
			require.NoError(t, err)
			require.True(t, plan.Empty(), "alias and API zone must produce an empty plan")

			zone = "Asia/Tokyo"
			state, ds = resource.RefreshWithoutUpgrade(t.Context(), state, meta)
			require.False(t, ds.HasError(), "%v", ds)
			plan, err = resource.Diff(t.Context(), state, cfg, meta)
			require.NoError(t, err)
			require.False(t, plan.Empty(), "a different remote zone must still be detected")
		})
	}
}
