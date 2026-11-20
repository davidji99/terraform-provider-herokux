package herokux

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestConnectMappingOwnership(t *testing.T) {
	const app = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	const conn = "38b6e411-1361-4adf-ace6-734c9a95513d"
	prefix := "/api/v3/connections/" + conn
	ids := map[string]string{}
	writes := []string{}
	var failImport bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		rows := []map[string]string{}
		for _, name := range slices.Sorted(maps.Keys(ids)) {
			rows = append(rows, map[string]string{"object_name": name, "id": ids[name]})
		}
		switch {
		case r.URL.Path == "/auth/"+app:
			fmt.Fprintf(w, `{"connections":[{"id":%q,"region_url":%q}]}`, conn, "http://"+r.Host)
		case r.URL.Path == prefix:
			_ = json.NewEncoder(w).Encode(map[string]any{"app_id": app, "addon_id": conn, "mappings": rows})
		case r.URL.Path == prefix+"/actions/export":
			for _, row := range rows {
				delete(row, "id")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"mappings": rows})
		case r.URL.Path == prefix+"/actions/import":
			if failImport {
				w.WriteHeader(422)
				return
			}
			var body struct{ Mappings []map[string]string }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				return
			}
			for _, m := range body.Mappings {
				name := m["object_name"]
				writes = append(writes, "import "+name)
				if ids[name] == "" {
					ids[name] = "id-" + name
				}
			}
			fmt.Fprint(w, "{}")
		case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/api/v3/mappings/"):
			id := strings.TrimPrefix(r.URL.Path, "/api/v3/mappings/")
			writes = append(writes, "delete "+id)
			for k, v := range ids {
				if v == id {
					delete(ids, k)
				}
			}
			fmt.Fprint(w, "{}")
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(400)
		}
	}))
	defer srv.Close()
	client, err := api.New(config.ConnectCentralBaseURL(srv.URL))
	require.NoError(t, err)
	meta := &Config{API: client}
	r := resourceHerokuxConnectMappings()
	var state *terraform.InstanceState
	apply := func(wantError bool, names ...string) {
		t.Helper()
		rows := []map[string]string{}
		for _, name := range names {
			rows = append(rows, map[string]string{"object_name": name})
		}
		document, err := json.Marshal(map[string]any{"mappings": rows})
		require.NoError(t, err)
		args := map[string]any{"app_id": app, "connect_id": conn, "mappings": string(document)}
		diff, err := r.Diff(t.Context(), state, terraform.NewResourceConfigRaw(args), meta)
		require.NoError(t, err)
		require.False(t, diff.Empty(), "expected adoption or mapping change")
		next, ds := r.Apply(t.Context(), state, diff, meta)
		require.Equal(t, wantError, ds.HasError(), "%v", ds)
		if !wantError {
			state = next
		}
	}
	refresh := func() {
		t.Helper()
		next, ds := r.RefreshWithoutUpgrade(t.Context(), state, meta)
		require.False(t, ds.HasError(), "%v", ds)
		state = next
	}
	destroy := func() {
		t.Helper()
		_, ds := r.Apply(t.Context(), state, &terraform.InstanceDiff{Destroy: true}, meta)
		require.False(t, ds.HasError(), "%v", ds)
		state = nil
	}
	adopt := func() {
		t.Helper()
		d := r.Data(nil)
		d.SetId(app + ":" + conn)
		entries, err := r.Importer.StateContext(t.Context(), d, meta)
		require.NoError(t, err)
		state = entries[0].State()
	}

	apply(false, "Account", "Lead")
	ids["Contact"] = "external"
	refresh()
	require.NotContains(t, state.Attributes["mappings"], "Contact")
	failImport = true
	apply(true, "Account")
	failImport = false
	require.Contains(t, ids, "Lead")
	apply(false, "Account")
	require.NotContains(t, ids, "Lead")
	require.Contains(t, ids, "Contact")
	destroy()
	require.Equal(t, map[string]string{"Contact": "external"}, ids)

	adopt()
	apply(false, "Contact") // The export already matches HCL. Adoption still needs an apply.
	require.Equal(t, "external", state.Attributes["mapping_data.Contact"])
	state = nil

	ids["Account"] = "old"
	adopt()
	destroy()
	require.Contains(t, ids, "Account")
	adopt()
	apply(false, "Account")
	require.Equal(t, "old", state.Attributes["mapping_data.Account"])
	ids["Account"] = "replacement"
	refresh()
	writes = nil
	apply(true, "Account")
	require.Empty(t, writes)
	destroy()
	require.Empty(t, writes)
	require.Equal(t, "replacement", ids["Account"])

	adopt()
	// Emulate an upgraded instance whose computed map used to include every remote mapping.
	delete(state.Attributes, "ownership_recorded")
	state.Attributes["mapping_data.%"] = "2"
	state.Attributes["mapping_data.Contact"] = "external"
	state.Attributes["mapping_data.Account"] = "replacement"
	legacy := state.DeepCopy()
	destroy()
	require.Empty(t, writes)
	state = legacy
	refresh()
	apply(false, "Account")
	destroy()
	require.Equal(t, map[string]string{"Contact": "external"}, ids)
}
