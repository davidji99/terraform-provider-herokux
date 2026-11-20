package herokux

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

type redisNotificationFixture struct {
	mu     sync.Mutex
	meta   *Config
	remote string
	posted []string
}

func newRedisNotificationFixture(t *testing.T) *redisNotificationFixture {
	t.Helper()
	f := &redisNotificationFixture{remote: "Ex"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if req.Method == http.MethodPatch {
			var payload map[string]any
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Error(err)
				return
			}
			if v, ok := payload["notify_keyspace_events"].(string); ok {
				f.posted = append(f.posted, v)
				f.remote = v
				if v == "KEA" {
					f.remote = "AKE"
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"notify_keyspace_events": map[string]string{"value": f.remote}})
	}))
	t.Cleanup(srv.Close)
	client, err := api.New(config.RedisBaseURL(srv.URL))
	require.NoError(t, err)
	f.meta = &Config{API: client}
	return f
}

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
	require.False(t, equivalentRedisNotificationFlags("En", "E"))
	f := newRedisNotificationFixture(t)
	meta := f.meta
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
	require.Empty(t, f.posted)
	for _, value := range []string{"KEA", "", "KEA", "disabled"} {
		apply(&value)
	}
	require.Equal(t, []string{"KEA", "", "KEA", ""}, f.posted)
}

// The dependent resource must receive exactly the string in the saved plan.
func TestRedisNotificationTerraformCLI(t *testing.T) {
	f := newRedisNotificationFixture(t)
	dir := t.TempDir()
	run := localTerraformCLI(t, dir, f.meta)
	const config = `terraform {
  required_providers { herokux = { source = "davidji99/herokux" } }
}
resource "herokux_redis_config" "test" {
  redis_id = "redis"
  notify_keyspace_events = "KEA"
}
resource "terraform_data" "consumer" { input = herokux_redis_config.test.notify_keyspace_events }
output "flags" { value = terraform_data.consumer.input }`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte(config), 0600))
	run(0, "plan", "-out=saved")
	run(0, "apply", "saved")
	require.Equal(t, "KEA", string(run(0, "output", "-raw", "flags")))
	run(0, "plan", "-detailed-exitcode")
	f.mu.Lock()
	before := f.remote
	f.remote = "Ex"
	f.mu.Unlock()
	require.Equal(t, "AKE", before)
	run(2, "plan", "-detailed-exitcode")
}
