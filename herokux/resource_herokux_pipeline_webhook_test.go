package herokux

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/davidji99/terraform-provider-herokux/api/platform"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineWebhookLifecycle(t *testing.T) {
	const pipelineID = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	const webhookID = "38b6e411-1361-4adf-ace6-734c9a95513d"
	const webhookURL = "https://example.com/webhook"
	const compositeID = pipelineID + ":" + webhookID
	collection := "/pipelines/" + pipelineID + "/webhooks"
	item := collection + "/" + webhookID

	type request struct {
		method, path string
		body         map[string]interface{}
	}
	requests := make(chan request, 16)
	remote := map[string]interface{}{
		"id": webhookID, "pipeline": map[string]string{"id": pipelineID, "name": "deploy"},
		"url": webhookURL, "level": "notify", "include": []string{"api:build"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, platform.WebhooksAcceptHeader, r.Header.Get("Accept"))
		assert.Equal(t, "Bearer provider-token", r.Header.Get("Authorization"))
		var body map[string]interface{}
		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
			if err := json.NewDecoder(r.Body).Decode(&body); !assert.NoError(t, err) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if body["secret"] == "rejected-secret" {
				requests <- request{r.Method, r.URL.Path, body}
				http.Error(w, "rejected-secret", http.StatusUnprocessableEntity)
				return
			}
			for _, key := range []string{"url", "level", "include"} {
				if value, ok := body[key]; ok {
					remote[key] = value
				}
			}
		}
		requests <- request{r.Method, r.URL.Path, body}
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			if body["secret"] == nil {
				w.Header().Set(platform.HerokuWebhookSecret, "generated-secret")
			}
			w.WriteHeader(http.StatusCreated)
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		assert.NoError(t, json.NewEncoder(w).Encode(remote))
	}))
	defer server.Close()
	expect := func(method, path string, body map[string]interface{}) {
		t.Helper()
		select {
		case actual := <-requests:
			require.Equal(t, request{method, path, body}, actual)
		default:
			t.Fatalf("missing %s %s", method, path)
		}
	}

	client, err := api.New(config.PlatformBaseURL(server.URL), config.APIToken("provider-token"))
	require.NoError(t, err)
	meta := &Config{API: client}
	res := New().ResourcesMap["herokux_pipeline_webhook"]
	ctx := t.Context()
	var state *terraform.InstanceState
	args := map[string]interface{}{
		"pipeline_id": pipelineID, "url": webhookURL, "level": "notify", "event_types": []interface{}{"api:build"},
	}
	plan := func() *terraform.InstanceDiff {
		t.Helper()
		diff, diffErr := res.Diff(ctx, state, terraform.NewResourceConfigRaw(args), meta)
		require.NoError(t, diffErr)
		return diff
	}
	apply := func(replace bool) {
		t.Helper()
		diff := plan()
		require.NotNil(t, diff)
		if state != nil {
			require.Equal(t, replace, diff.RequiresNew())
		}
		next, diagnostics := res.Apply(ctx, state, diff, meta)
		require.False(t, diagnostics.HasError(), "%v", diagnostics)
		state = next
		require.Equal(t, compositeID, state.ID)
	}

	apply(false)
	expect("POST", collection, map[string]interface{}{"url": webhookURL, "level": "notify", "include": []interface{}{"api:build"}})
	require.Equal(t, "deploy", state.Attributes["pipeline_name"])
	require.Equal(t, "generated-secret", state.Attributes["signing_secret"])
	state, diagnostics := res.RefreshWithoutUpgrade(ctx, state, meta)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
	expect("GET", item, nil)
	require.Equal(t, "generated-secret", state.Attributes["signing_secret"])
	require.True(t, plan().Empty())

	// Import by name. A later update must leave the receiver's unknown credentials alone.
	importID := "deploy:" + webhookID
	data := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
	data.SetId(importID)
	imported, err := res.Importer.StateContext(ctx, data, meta)
	require.NoError(t, err)
	state, diagnostics = res.RefreshWithoutUpgrade(ctx, imported[0].State(), meta)
	require.False(t, diagnostics.HasError(), "%v", diagnostics)
	expect("GET", "/pipelines/deploy/webhooks/"+webhookID, nil)
	require.Equal(t, compositeID, state.ID)
	require.Empty(t, state.Attributes["signing_secret"])
	require.True(t, plan().Empty())

	args["url"] = webhookURL + "/new"
	apply(false)
	expect("PATCH", item, map[string]interface{}{"url": args["url"]})
	expect("GET", item, nil)

	args["level"] = "sync"
	args["event_types"] = []interface{}{"api:release"}
	args["secret"] = "configured-secret"
	args["authorization"] = "Bearer receiver"
	apply(false)
	expect("PATCH", item, map[string]interface{}{
		"level": "sync", "include": []interface{}{"api:release"}, "secret": args["secret"], "authorization": args["authorization"],
	})
	expect("GET", item, nil)
	delete(args, "authorization")
	args["secret"] = "rotated-secret"
	apply(false)
	expect("PATCH", item, map[string]interface{}{"authorization": nil, "secret": "rotated-secret"})
	expect("GET", item, nil)

	delete(args, "secret")
	apply(true)
	expect("DELETE", item, nil)
	expect("POST", collection, map[string]interface{}{"url": args["url"], "level": "sync", "include": []interface{}{"api:release"}})
	require.Equal(t, "generated-secret", state.Attributes["signing_secret"])
	args["secret"] = "replacement-secret"
	apply(false)
	expect("PATCH", item, map[string]interface{}{"secret": args["secret"]})
	expect("GET", item, nil)
	require.Empty(t, state.Attributes["signing_secret"])
	require.True(t, plan().Empty())
	args["secret"] = "rejected-secret"
	state, diagnostics = res.Apply(ctx, state, plan(), meta)
	require.True(t, diagnostics.HasError())
	require.NotContains(t, fmt.Sprint(diagnostics), "rejected-secret")
	require.Equal(t, "replacement-secret", state.Attributes["secret"])
	expect("PATCH", item, map[string]interface{}{"secret": "rejected-secret"})
	require.False(t, plan().Empty(), "a failed credential rotation should remain pending")
	require.Empty(t, requests)
}

func TestPipelineWebhookErrors(t *testing.T) {
	res := resourceHerokuxPipelineWebhook()
	for _, status := range []int{http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "secret-in-error-body", status)
			}))
			defer server.Close()
			client, err := api.New(config.PlatformBaseURL(server.URL))
			require.NoError(t, err)
			meta := &Config{API: client}
			const id = "deploy:38b6e411-1361-4adf-ace6-734c9a95513d"
			for _, run := range []schema.ReadContextFunc{res.ReadContext, schema.ReadContextFunc(res.DeleteContext)} {
				d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
				d.SetId(id)
				diagnostics := run(t.Context(), d, meta)
				if status == http.StatusNotFound {
					require.False(t, diagnostics.HasError(), "%v", diagnostics)
					require.Empty(t, d.Id())
				} else {
					require.True(t, diagnostics.HasError())
					require.Equal(t, id, d.Id())
					require.NotContains(t, fmt.Sprint(diagnostics), "secret-in-error-body")
				}
			}
		})
	}
}

func TestAccHerokuxPipelineWebhook(t *testing.T) {
	pipelineID := testAccConfig.GetPipelineIDorSkip(t)
	const address = "herokux_pipeline_webhook.test"
	initial := fmt.Sprintf(`
resource "herokux_pipeline_webhook" "test" {
  pipeline_id = %q
  url         = "https://example.com/heroku"
  level       = "notify"
  event_types = ["api:build"]
}`, pipelineID)
	updated := strings.Replace(initial, "notify", "sync", 1)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) }, Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: initial, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(address, "pipeline_id", pipelineID),
				resource.TestCheckResourceAttrSet(address, "pipeline_name"),
				resource.TestCheckResourceAttrSet(address, "signing_secret"),
			)},
			{Config: updated, Check: resource.TestCheckResourceAttr(address, "level", "sync")},
			{ResourceName: address, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"signing_secret"}},
			{Config: updated, PlanOnly: true},
		},
	})
}
