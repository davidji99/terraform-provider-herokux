package herokux

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestAppWebhookAuthorization(t *testing.T) {
	const appID = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	const signing = "signing-test-value"
	const header = "Bearer receiver-test-value"
	var logs bytes.Buffer
	saved := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(saved)
	var writes []map[string]any
	remote := map[string]any{"id": "subscription", "app": map[string]string{"id": appID}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				return
			}
			writes = append(writes, body)
			if body["authorization"] == "rejected-test-value" {
				http.Error(w, "rejected-test-value", 422)
				return
			}
			for _, key := range []string{"url", "level", "include", "name"} {
				if v, ok := body[key]; ok {
					remote[key] = v
				}
			}
		}
		_ = json.NewEncoder(w).Encode(remote)
	}))
	defer server.Close()
	client, err := api.New(config.PlatformBaseURL(server.URL))
	require.NoError(t, err)
	r := resourceHerokuxAppWebhook()
	meta := &Config{API: client}
	var state *terraform.InstanceState
	args := map[string]any{"app_id": appID, "url": "https://example.test/hook", "event_types": []any{"api:release"}, "level": "notify", "secret": signing, "authorization": header}
	apply := func(fail bool) map[string]any {
		t.Helper()
		diff, err := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(args), meta)
		require.NoError(t, err)
		next, ds := r.Apply(t.Context(), state, diff, meta)
		require.Equal(t, fail, ds.HasError(), "%v", ds)
		if !fail {
			state = next
		} else {
			require.NotContains(t, ds[0].Detail, "rejected-test-value")
		}
		require.Len(t, writes, 1)
		body := writes[0]
		writes = nil
		return body
	}
	create := apply(false)
	require.Equal(t, header, create["authorization"])
	require.Equal(t, signing, create["secret"])
	args["level"] = "sync"
	require.Equal(t, map[string]any{"level": "sync"}, apply(false))
	args["authorization"] = "changed-test-value"
	require.Equal(t, map[string]any{"authorization": "changed-test-value"}, apply(false))
	delete(args, "authorization")
	require.Equal(t, map[string]any{"authorization": nil}, apply(false))
	args["authorization"] = "rejected-test-value"
	apply(true)
	for _, value := range []string{signing, header, "changed-test-value", "rejected-test-value"} {
		require.NotContains(t, logs.String(), value)
	}
}
