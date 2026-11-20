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

func TestRedisNotificationFlags(t *testing.T) {
	r := resourceHerokuxRedisConfig()
	field := r.Schema["notify_keyspace_events"]
	for _, v := range []string{"K", "KEA", "E$", "", "disabled"} {
		_, errs := field.ValidateFunc(v, "notify_keyspace_events")
		require.Empty(t, errs)
	}
	_, errs := field.ValidateFunc("Ex?", "notify_keyspace_events")
	require.NotEmpty(t, errs)
	require.True(t, equivalentRedisNotificationFlags("AKEm", "g$lshztxemKE"))
	var posted []string
	remote := "Ex"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPatch {
			var payload map[string]any
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Error(err)
				return
			}
			if v, ok := payload["notify_keyspace_events"].(string); ok {
				posted = append(posted, v)
				remote = v
				if v == "KEA" {
					remote = "AKE"
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"notify_keyspace_events": map[string]string{"value": remote}})
	}))
	defer srv.Close()
	client, err := api.New(config.RedisBaseURL(srv.URL))
	require.NoError(t, err)
	meta := &Config{API: client}
	var state *terraform.InstanceState
	apply := func(flag *string) {
		t.Helper()
		args := map[string]any{"redis_id": "redis"}
		if flag != nil {
			args["notify_keyspace_events"] = *flag
		}
		cfg := terraform.NewResourceConfigRaw(args)
		diff, err := r.Diff(t.Context(), state, cfg, meta)
		require.NoError(t, err)
		next, ds := r.Apply(t.Context(), state, diff, meta)
		require.False(t, ds.HasError(), "%v", ds)
		state = next
		diff, err = r.Diff(t.Context(), state, cfg, meta)
		require.NoError(t, err)
		require.True(t, diff.Empty(), "%v", diff)
	}
	apply(nil)
	require.Empty(t, posted)
	for _, value := range []string{"KEA", "", "KEA", "disabled"} {
		apply(&value)
	}
	require.Equal(t, []string{"KEA", "", "KEA", ""}, posted)
}
