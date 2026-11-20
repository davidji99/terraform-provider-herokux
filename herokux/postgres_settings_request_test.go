package herokux

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestPostgresSettingsPatch(t *testing.T) {
	for _, tc := range []struct {
		name             string
		initial, updated map[string]any
	}{
		{"omitted", map[string]any{}, nil},
		{"zero on create", map[string]any{"log_connections": false, "log_lock_waits": false, "log_min_duration_statement": 0}, nil},
		{"zero on update", map[string]any{"log_connections": true, "log_lock_waits": true, "log_min_duration_statement": 500}, map[string]any{"log_connections": false, "log_lock_waits": false, "log_min_duration_statement": 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var sent []map[string]any
			values := map[string]any{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPatch {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
						return
					}
					sent = append(sent, body)
					for k, v := range body {
						values[k] = map[string]any{"value": v}
					}
				}
				_ = json.NewEncoder(w).Encode(values)
			}))
			defer server.Close()
			client, err := api.New(config.PostgresBaseURL(server.URL))
			require.NoError(t, err)
			resource := resourceHerokuxPostgresSettings()
			var state *terraform.InstanceState
			apply := func(values map[string]any) {
				t.Helper()
				args := map[string]any{"postgres_id": "db"}
				for k, v := range values {
					args[k] = v
				}
				meta := &Config{API: client}
				diff, err := resource.Diff(t.Context(), state, terraform.NewResourceConfigRaw(args), meta)
				require.NoError(t, err)
				next, ds := resource.Apply(t.Context(), state, diff, meta)
				require.False(t, ds.HasError(), "%v", ds)
				state = next
				require.Len(t, sent, 1)
				wanted, err := json.Marshal(values)
				require.NoError(t, err)
				actual, err := json.Marshal(sent[0])
				require.NoError(t, err)
				require.JSONEq(t, string(wanted), string(actual))
				sent = nil
			}
			apply(tc.initial)
			if tc.updated != nil {
				apply(tc.updated)
			}
		})
	}
}
