package herokux

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/davidji99/terraform-provider-herokux/api/platform"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestContainerReleaseDrift(t *testing.T) {
	const app = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	const process = "web"
	a := "sha256:" + strings.Repeat("a", 64)
	b := "sha256:" + strings.Repeat("b", 64)
	remote, etag := a, `"version-a"`
	getStatus, patchStatus := 200, 200
	removes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/apps/"+app && r.Method == "GET" {
			fmt.Fprintf(w, `{"id":%q}`, app)
			return
		}
		if r.URL.Path != "/apps/"+app+"/formation/web" || r.Header.Get("Accept") != platform.DockerReleasesAcceptHeader {
			t.Errorf("unexpected request %v", r.URL.Path)
			w.WriteHeader(400)
			return
		}
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("ETag", etag)
			w.WriteHeader(getStatus)
			fmt.Fprintf(w, `{"app":{"id":%q},"type":"web","docker_image":{"id":%q}}`, app, remote)
		case http.MethodPatch:
			removes++
			if r.Header.Get("If-Match") != etag || etag == "" {
				t.Error("missing removal precondition")
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if value, ok := body["docker_image"]; !ok || value != nil {
				t.Error("expected a null docker_image")
			}
			w.WriteHeader(patchStatus)
		default:
			t.Error(r.Method)
		}
	}))
	defer srv.Close()
	client, err := api.New(config.PlatformBaseURL(srv.URL))
	require.NoError(t, err)
	platformClient := heroku.NewService(srv.Client())
	platformClient.URL = srv.URL
	meta := &Config{API: client, PlatformAPI: platformClient}
	r := resourceHerokuxAppContainerRelease()
	d := r.Data(nil)
	d.SetId(app + "|" + a + "|web")
	entries, err := r.Importer.StateContext(t.Context(), d, meta)
	require.NoError(t, err)
	d = entries[0]
	require.Equal(t, a, d.Get("last_applied_image_id"))
	// Old state only recorded image_id. Preserve it when the observed image has changed.
	require.NoError(t, d.Set("last_applied_image_id", ""))
	remote = b
	ds := r.ReadContext(t.Context(), d, meta)
	require.False(t, ds.HasError(), "%v", ds)
	require.Equal(t, a, d.Get("last_applied_image_id"))
	require.Equal(t, b, d.Get("image_id"))
	diff, err := r.Diff(t.Context(), d.State(), terraform.NewResourceConfigRaw(map[string]any{"app_id": app, "process_type": process, "image_id": a}), meta)
	require.NoError(t, err)
	require.Equal(t, b, diff.Attributes["image_id"].Old)
	require.Equal(t, a, diff.Attributes["image_id"].New)
	for _, bad := range []string{b, "", "short-digest"} {
		remote = bad
		ds = r.DeleteContext(t.Context(), d, meta)
		require.True(t, ds.HasError())
		require.NotEmpty(t, d.Id())
		require.Zero(t, removes)
	}
	getStatus = 403
	ds = r.DeleteContext(t.Context(), d, meta)
	require.True(t, ds.HasError())
	require.Zero(t, removes)
	getStatus = 200
	remote = a
	etag = ""
	ds = r.DeleteContext(t.Context(), d, meta)
	require.True(t, ds.HasError())
	require.Zero(t, removes)
	etag = `"v3"`
	patchStatus = 412
	ds = r.DeleteContext(t.Context(), d, meta)
	require.True(t, ds.HasError())
	require.NotEmpty(t, d.Id())
	require.Equal(t, 1, removes)
	patchStatus = 200
	ds = r.DeleteContext(t.Context(), d, meta)
	require.False(t, ds.HasError(), "%v", ds)
	require.Empty(t, d.Id())
	require.Equal(t, 2, removes)
	d.SetId(app + ":web")
	getStatus = 404
	ds = r.DeleteContext(t.Context(), d, meta)
	require.False(t, ds.HasError(), "%v", ds)
	require.Empty(t, d.Id())
	require.Equal(t, 2, removes)
}
