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

func TestDataConnectorInPlaceUpdates(t *testing.T) {
	const source = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	const store = "38b6e411-1361-4adf-ace6-734c9a95513d"
	remote := map[string]any{"uuid": "connector", "name": "stream", "status": "paused", "postgres_addon": map[string]string{"uuid": source}, "kafka_addon": map[string]string{"uuid": store}, "tables": []string{"public.a"}, "excluded_columns": []string{"public.a.x"}, "settings": map[string]string{"mode": "original"}}
	var writes []map[string]any
	failRead := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/data/cdc/v0/connectors/connector" {
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(400)
			return
		}
		switch r.Method {
		case http.MethodPatch:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				return
			}
			writes = append(writes, body)
			for k, v := range body {
				remote[k] = v
			}
		case http.MethodGet:
			if failRead {
				failRead = false
				w.WriteHeader(403)
				return
			}
		default:
			t.Errorf("unexpected %s", r.Method)
			w.WriteHeader(400)
			return
		}
		_ = json.NewEncoder(w).Encode(remote)
	}))
	defer srv.Close()
	client, err := api.New(config.PostgresBaseURL(srv.URL))
	require.NoError(t, err)
	r := resourceHerokuxDataConnector()
	meta := &Config{API: client, DataConnectorSettingsUpdateVerifyTimeout: 1}
	d := r.Data(nil)
	d.SetId("connector")
	ds := r.ReadContext(t.Context(), d, meta)
	require.False(t, ds.HasError(), "%v", ds)
	state := d.State()
	args := map[string]any{"source_id": source, "store_id": store, "tables": []any{"public.a", "public.b"}, "excluded_columns": []any{}}
	plan := func() *terraform.InstanceDiff {
		diff, err := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(args), meta)
		require.NoError(t, err)
		return diff
	}
	diff := plan()
	require.False(t, diff.RequiresNew())
	failRead = true
	_, ds = r.Apply(t.Context(), state, diff, meta)
	require.True(t, ds.HasError())
	require.Len(t, writes, 1)
	require.Equal(t, map[string]any{"tables": []any{"public.a", "public.b"}, "excluded_columns": []any{}}, writes[0])
	state, ds = r.RefreshWithoutUpgrade(t.Context(), state, meta)
	require.False(t, ds.HasError(), "%v", ds)
	require.True(t, plan().Empty(), "partial apply should converge after refresh")
	args["tables"] = []any{"public.b"}
	args["excluded_columns"] = []any{"public.b.x"}
	diff = plan()
	require.False(t, diff.RequiresNew())
	state, ds = r.Apply(t.Context(), state, diff, meta)
	require.False(t, ds.HasError(), "%v", ds)
	require.Equal(t, "connector", state.ID)
	require.Equal(t, map[string]any{"tables": []any{"public.b"}, "excluded_columns": []any{"public.b.x"}}, writes[1])
	require.True(t, plan().Empty())
	require.Equal(t, "paused", remote["status"])
	require.Equal(t, map[string]string{"mode": "original"}, remote["settings"])
}
