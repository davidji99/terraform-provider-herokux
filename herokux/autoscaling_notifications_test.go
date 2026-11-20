package herokux

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoscalingNotificationUpdates(t *testing.T) {
	const app = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	const formation = "/apps/" + app + "/formation/web"
	monitor := map[string]any{"id": "monitor", "app_id": app, "process_type": "web", "is_active": true, "min_quantity": 1, "max_quantity": 2, "value": 1000, "notification_channels": []string{"app"}}
	var patches []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == formation+"/monitors/monitor":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			patches = append(patches, body)
			for k, v := range body {
				monitor[k] = v
			}
			w.WriteHeader(http.StatusAccepted)
		case r.Method == http.MethodGet && r.URL.Path == formation+"/monitors/monitor":
			assert.NoError(t, json.NewEncoder(w).Encode(monitor))
		case r.Method == http.MethodGet && r.URL.Path == formation:
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]string{"size": "Performance-M"}))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(400)
		}
	}))
	defer srv.Close()
	client, err := api.New(config.MetricsBaseURL(srv.URL))
	require.NoError(t, err)
	platform := heroku.NewService(srv.Client())
	platform.URL = srv.URL
	meta := &Config{API: client, PlatformAPI: platform}
	r := resourceHerokuxFormationAutoscaling()
	d := r.Data(nil)
	d.SetId(app + ":web:monitor")
	ds := r.ReadContext(t.Context(), d, meta)
	require.False(t, ds.HasError(), "%v", ds)
	state := d.State()
	args := map[string]any{"app_id": app, "process_type": "web", "dyno_size": "Performance-M", "is_active": true, "min_quantity": 1, "max_quantity": 3, "desired_p95_response_time": 1000, "notification_channels": []any{"app"}}
	apply := func() map[string]any {
		t.Helper()
		cfg := terraform.NewResourceConfigRaw(args)
		diff, err := r.Diff(t.Context(), state, cfg, meta)
		require.NoError(t, err)
		require.False(t, diff.Empty())
		state, ds = r.Apply(t.Context(), state, diff, meta)
		require.False(t, ds.HasError(), "%v", ds)
		require.Len(t, patches, 1)
		body := patches[0]
		patches = nil
		return body
	}
	patch := apply()
	require.NotContains(t, patch, "notification_channels", "updating the dyno limit should leave notification recipients alone")
	require.Equal(t, "app", state.Attributes["notification_channels.0"])
	args["notification_channels"] = []any{}
	require.Equal(t, []any{}, apply()["notification_channels"])
	require.Equal(t, "0", state.Attributes["notification_channels.#"])
	args["notification_channels"] = []any{"app"}
	require.Equal(t, []any{"app"}, apply()["notification_channels"])
}
