package herokux

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestPlatformInventory(t *testing.T) {
	for _, tc := range []struct {
		path, attribute string
		resource        *schema.Resource
		args            map[string]any
	}{
		{"/addons", "addons", dataSourceHerokuxAddons(), map[string]any{"addon_name_regex": "^match"}},
		{"/apps/app/addons", "addons", dataSourceHerokuxAppAddons(), map[string]any{"app_id": "app", "addon_service_name": "match"}},
		{"/apps", "apps", dataSourceHerokuxSpaceApps(), map[string]any{"space_regex": "^match"}},
	} {
		t.Run(tc.path, func(t *testing.T) {
			secondStatus := 200
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.path {
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(404)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				name := "other"
				switch r.Header.Get("Range") {
				case "":
					w.Header().Set("Next-Range", "id opaque..")
					w.WriteHeader(206)
				case "id opaque..":
					name = "match"
					w.WriteHeader(secondStatus)
				default:
					t.Error("unexpected range")
					w.WriteHeader(400)
					return
				}
				fmt.Fprintf(w, `[{"id":%q,"name":%q,"space":{"name":%q},"addon_service":{"name":%q}}]`, name, name, name, name)
			}))
			defer srv.Close()
			client, err := api.New(config.PlatformBaseURL(srv.URL))
			require.NoError(t, err)
			d := schema.TestResourceDataRaw(t, tc.resource.Schema, tc.args)
			read := func() bool { return tc.resource.ReadContext(t.Context(), d, &Config{API: client}).HasError() }
			require.False(t, read())
			switch value := d.Get(tc.attribute).(type) {
			case []any:
				require.Len(t, value, 1)
				require.Equal(t, "match", value[0].(map[string]any)["name"])
			case map[string]any:
				require.Equal(t, map[string]any{"match": "match"}, value)
			default:
				t.Fatalf("unexpected inventory: %T", value)
			}
			id := d.Id()
			require.NotEmpty(t, id)
			require.False(t, read())
			require.Equal(t, id, d.Id())
			secondStatus = 403
			require.True(t, read())
		})
	}
}

func TestProviderDataEndpointRouting(t *testing.T) {
	requests := []string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/data/graphql":
			fmt.Fprint(w, `{"data":{"clip":{"id":"ok"}}}`)
		case "/pg/data/kafka/v0/clusters/cluster":
			fmt.Fprint(w, `{"name":"cluster"}`)
		default:
			t.Errorf("unexpected URL %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	c := NewConfig()
	require.NoError(t, c.applySchema(schema.TestResourceDataRaw(t, New().Schema, map[string]any{"data_api_url": srv.URL + "/data", "postgres_api_url": srv.URL + "/pg"})))
	require.NoError(t, c.initializeAPI())
	clip, _, err := c.API.Data.GetPostgresDataclip("clip")
	require.NoError(t, err)
	require.Equal(t, "ok", clip.GetID())
	cluster, _, err := c.API.Kafka.Get("cluster")
	require.NoError(t, err)
	require.Equal(t, "cluster", cluster.GetName())
	require.Len(t, requests, 2)
	_, err = api.New(config.DataBaseURL(""))
	require.Error(t, err)
}
