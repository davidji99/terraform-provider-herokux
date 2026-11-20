package herokux

import (
	"encoding/json"
	"fmt"
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
		{"minimum error severity", map[string]any{"log_min_error_statement": "panic"}, map[string]any{"log_min_error_statement": "error"}},
		{"leave tracking unmanaged", map[string]any{"log_min_error_statement": "panic", "track_functions": "all"}, map[string]any{"log_min_error_statement": "log"}},
		{"function tracking", map[string]any{"track_functions": "pl"}, map[string]any{"track_functions": "none"}},
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
				for key, value := range values {
					require.Equal(t, fmt.Sprint(value), state.Attributes[key], key)
				}
				sent = nil
			}
			apply(tc.initial)
			if tc.updated != nil {
				apply(tc.updated)
			}
		})
	}
}

func TestPostgresSettingsStringOptions(t *testing.T) {
	r := resourceHerokuxPostgresSettings()
	for _, tc := range []struct {
		key            string
		valid, invalid []string
	}{
		{"log_min_error_statement", []string{"error", "log", "fatal", "panic"}, []string{"warning", "ERROR", ""}},
		{"track_functions", []string{"none", "pl", "all"}, []string{"sql", "ALL", ""}},
	} {
		t.Run(tc.key, func(t *testing.T) {
			for _, value := range tc.valid {
				ds := r.Validate(terraform.NewResourceConfigRaw(map[string]any{"postgres_id": "db", tc.key: value}))
				require.False(t, ds.HasError(), "value %q: %v", value, ds)
			}
			for _, value := range tc.invalid {
				ds := r.Validate(terraform.NewResourceConfigRaw(map[string]any{"postgres_id": "db", tc.key: value}))
				require.True(t, ds.HasError(), "value %q", value)
			}
		})
	}
}

func TestPostgresSettingsImportAndRefresh(t *testing.T) {
	// Start with settings configured through the CLI.
	remote := map[string]map[string]string{"log_min_error_statement": {"value": "log"}, "track_functions": {"value": "all"}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/postgres/v0/databases/db/config" {
			t.Errorf("unexpected %s %s", req.Method, req.URL.Path)
			w.WriteHeader(400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(remote)
	}))
	defer srv.Close()
	client, err := api.New(config.PostgresBaseURL(srv.URL))
	require.NoError(t, err)
	meta := &Config{API: client}
	r := resourceHerokuxPostgresSettings()
	d := r.Data(nil)
	d.SetId("db")
	imported, err := r.Importer.StateContext(t.Context(), d, meta)
	require.NoError(t, err)
	require.Len(t, imported, 1)
	state := imported[0].State()
	require.Equal(t, "log", state.Attributes["log_min_error_statement"])
	require.Equal(t, "all", state.Attributes["track_functions"])
	cfg := terraform.NewResourceConfigRaw(map[string]any{"postgres_id": "db", "log_min_error_statement": "log", "track_functions": "all"})
	diff, err := r.Diff(t.Context(), state, cfg, meta)
	require.NoError(t, err)
	require.True(t, diff.Empty())
	remote["log_min_error_statement"]["value"] = "fatal"
	remote["track_functions"]["value"] = "pl"
	refreshed, ds := r.RefreshWithoutUpgrade(t.Context(), state, meta)
	require.False(t, ds.HasError(), "%v", ds)
	diff, err = r.Diff(t.Context(), refreshed, cfg, meta)
	require.NoError(t, err)
	require.False(t, diff.RequiresNew())
	require.Contains(t, diff.Attributes, "log_min_error_statement")
	require.Equal(t, "fatal", diff.Attributes["log_min_error_statement"].Old)
	require.Contains(t, diff.Attributes, "track_functions")
	require.Equal(t, "pl", diff.Attributes["track_functions"].Old)

	// Older responses can omit either setting.
	delete(remote, "log_min_error_statement")
	delete(remote, "track_functions")
	sparse, ds := r.RefreshWithoutUpgrade(t.Context(), refreshed, meta)
	require.False(t, ds.HasError(), "%v", ds)
	require.Empty(t, sparse.Attributes["log_min_error_statement"])
	require.Empty(t, sparse.Attributes["track_functions"])
	diff, err = r.Diff(t.Context(), sparse, terraform.NewResourceConfigRaw(map[string]any{"postgres_id": "db"}), meta)
	require.NoError(t, err)
	require.True(t, diff.Empty())
}
