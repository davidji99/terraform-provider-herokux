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

const (
	connectFixtureApp = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	connectFixtureID  = "38b6e411-1361-4adf-ace6-734c9a95513d"
)

type connectMappingFixture struct {
	ids        map[string]string
	writes     []string
	failImport bool
	meta       *Config
}

func newConnectMappingFixture(t *testing.T, ids map[string]string) *connectMappingFixture {
	t.Helper()
	f := &connectMappingFixture{ids: ids}
	const app, conn = connectFixtureApp, connectFixtureID
	prefix := "/api/v3/connections/" + conn
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		rows := []map[string]string{}
		for _, name := range slices.Sorted(maps.Keys(f.ids)) {
			rows = append(rows, map[string]string{"object_name": name, "id": f.ids[name]})
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
			if f.failImport {
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
				f.writes = append(f.writes, "import "+name)
				if f.ids[name] == "" {
					f.ids[name] = "id-" + name
				}
			}
			fmt.Fprint(w, "{}")
		case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/api/v3/mappings/"):
			id := strings.TrimPrefix(r.URL.Path, "/api/v3/mappings/")
			f.writes = append(f.writes, "delete "+id)
			for k, v := range f.ids {
				if v == id {
					delete(f.ids, k)
				}
			}
			fmt.Fprint(w, "{}")
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(400)
		}
	}))
	t.Cleanup(srv.Close)
	client, err := api.New(config.ConnectCentralBaseURL(srv.URL))
	require.NoError(t, err)
	f.meta = &Config{API: client}
	return f
}

func TestConnectMappingOwnership(t *testing.T) {
	const app, conn = connectFixtureApp, connectFixtureID
	f := newConnectMappingFixture(t, map[string]string{})
	meta := f.meta
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
		state = next
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
	f.ids["Contact"] = "external"
	refresh()
	require.NotContains(t, state.Attributes["mappings"], "Contact")
	f.failImport = true
	apply(true, "Account")
	f.failImport = false
	require.Contains(t, f.ids, "Lead")
	apply(false, "Account")
	require.NotContains(t, f.ids, "Lead")
	require.Contains(t, f.ids, "Contact")
	destroy()
	require.Equal(t, map[string]string{"Contact": "external"}, f.ids)

	adopt()
	require.JSONEq(t, emptyConnectMappings, state.Attributes["mappings"])
	f.failImport = true
	apply(true, "Contact")
	require.NotEqual(t, "true", state.Attributes["ownership_recorded"])
	destroy()
	require.Equal(t, "external", f.ids["Contact"])
	f.failImport = false
	adopt()
	apply(false, "Contact")
	require.Equal(t, "external", state.Attributes["mapping_data.Contact"])
	state = nil

	f.ids["Account"] = "old"
	adopt()
	destroy()
	require.Contains(t, f.ids, "Account")
	adopt()
	apply(false, "Account")
	require.Equal(t, "old", state.Attributes["mapping_data.Account"])
	f.ids["Account"] = "replacement"
	refresh()
	f.writes = nil
	apply(true, "Account")
	require.Empty(t, f.writes)
	destroy()
	require.Empty(t, f.writes)
	require.Equal(t, "replacement", f.ids["Account"])

	adopt()
	// Emulate an upgraded instance whose computed map used to include every remote mapping.
	delete(state.Attributes, "ownership_recorded")
	state.Attributes["mapping_data.%"] = "2"
	state.Attributes["mapping_data.Contact"] = "external"
	state.Attributes["mapping_data.Account"] = "replacement"
	legacy := state.DeepCopy()
	// An ID retained for reconfirmation can reject a same-name replacement,
	// but the failed adoption must not give destroy authority over either ID.
	state.Attributes["mapping_data.Account"] = "old"
	refresh()
	apply(true, "Account")
	require.NotEqual(t, "true", state.Attributes["ownership_recorded"])
	destroy()
	require.Empty(t, f.writes)
	state = legacy
	refresh()
	apply(false, "Account")
	destroy()
	require.Equal(t, map[string]string{"Contact": "external"}, f.ids)

	apply(false, "Lead")
	f.writes = nil
	apply(false) // Removing the last owned mapping must still delete it.
	require.Equal(t, []string{"delete id-Lead"}, f.writes)
	require.Equal(t, map[string]string{"Contact": "external"}, f.ids)
}
