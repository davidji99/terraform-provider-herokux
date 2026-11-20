package herokux

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShieldPrivateSpaceTeamTransfer(t *testing.T) {
	const spaceID = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	const oldTeam = "38b6e411-1361-4adf-ace6-734c9a95513d"
	const newTeam = "0f17fba1-269a-45fb-b62d-ebcc6e749987"
	const spacePath = "/spaces/" + spaceID
	const drain = "https://example.test/logs"

	for _, tc := range []struct {
		name                         string
		lookupStatus, transferStatus int
		rejectRename                 bool
	}{
		{"success", http.StatusOK, http.StatusOK, false},
		{"rejected transfer", http.StatusOK, http.StatusForbidden, false},
		{"destination lookup denied", http.StatusForbidden, http.StatusOK, false},
		{"transfer succeeds before rename fails", http.StatusOK, http.StatusOK, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			owner := map[string]string{"id": oldTeam, "name": "source-team"}
			space := map[string]any{"id": spaceID, "name": "billing", "region": map[string]string{"name": "virginia"}, "shield": true, "team": owner, "cidr": "10.0.0.0/16", "data_cidr": "10.1.0.0/16"}
			var writes []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method != http.MethodGet {
					writes = append(writes, r.Method+" "+r.URL.Path)
				}
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/teams/"+newTeam:
					w.WriteHeader(tc.lookupStatus)
					if tc.lookupStatus == 200 {
						_ = json.NewEncoder(w).Encode(map[string]string{"id": newTeam, "name": "destination-team"})
					} else {
						_, _ = w.Write([]byte(`{"id":"forbidden","message":"destination lookup denied"}`))
					}
				case r.Method == http.MethodPost && r.URL.Path == spacePath+"/transfer":
					var body map[string]string
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					assert.Equal(t, map[string]string{"new_owner": "destination-team"}, body)
					w.WriteHeader(tc.transferStatus)
					if tc.transferStatus != 200 {
						_, _ = w.Write([]byte(`{"id":"forbidden","message":"transfer denied"}`))
						return
					}
					owner["id"], owner["name"] = newTeam, "destination-team"
					_ = json.NewEncoder(w).Encode(space)
				case tc.rejectRename && r.Method == http.MethodPatch && r.URL.Path == spacePath:
					w.WriteHeader(http.StatusUnprocessableEntity)
					_, _ = w.Write([]byte(`{"id":"invalid","message":"rename rejected"}`))
				case r.Method == http.MethodGet && r.URL.Path == spacePath:
					_ = json.NewEncoder(w).Encode(space)
				case r.Method == http.MethodGet && r.URL.Path == spacePath+"/nat":
					_, _ = w.Write([]byte(`{"sources":["192.0.2.1"]}`))
				case r.Method == http.MethodGet && r.URL.Path == spacePath+"/log-drain":
					_ = json.NewEncoder(w).Encode(map[string]string{"url": drain})
				default:
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
					w.WriteHeader(400)
				}
			}))
			defer server.Close()
			apiClient, err := api.New(config.PlatformBaseURL(server.URL))
			require.NoError(t, err)
			platform := heroku.NewService(&http.Client{Transport: &heroku.Transport{Transport: server.Client().Transport}})
			platform.URL = server.URL
			res := resourceHerokuxShieldPrivateSpace()
			meta := &Config{API: apiClient, PlatformAPI: platform}
			d := res.Data(nil)
			d.SetId(spaceID)
			ds := res.ReadContext(t.Context(), d, meta)
			require.False(t, ds.HasError(), "%v", ds)
			before := d.State()
			args := map[string]any{"name": "billing", "region": "virginia", "log_drain": drain, "team_id": newTeam}
			if tc.name == "success" {
				args["team_id"] = strings.ToUpper(newTeam)
			}
			if tc.rejectRename {
				args["name"] = "new-name"
			}
			cfg := terraform.NewResourceConfigRaw(args)
			plan, err := res.Diff(t.Context(), before, cfg, meta)
			require.NoError(t, err)
			require.False(t, plan.Empty())
			require.False(t, plan.RequiresNew(), "changing team_id must update the same space")
			after, ds := res.Apply(t.Context(), before, plan, meta)
			require.NotNil(t, after)
			require.Equal(t, spaceID, after.ID)
			if tc.lookupStatus == 200 {
				expected := []string{"POST " + spacePath + "/transfer"}
				if tc.rejectRename {
					expected = append(expected, "PATCH "+spacePath)
				}
				require.Equal(t, expected, writes)
			} else {
				require.Empty(t, writes)
			}
			rejected := tc.lookupStatus != 200 || tc.transferStatus != 200
			require.Equal(t, rejected || tc.rejectRename, ds.HasError(), "%v", ds)
			require.Equal(t, "billing", after.Attributes["name"])
			if rejected {
				require.Equal(t, oldTeam, after.Attributes["team_id"])
				require.Equal(t, "source-team", after.Attributes["team_name"])
			} else {
				require.Equal(t, newTeam, after.Attributes["team_id"])
				require.Equal(t, "destination-team", after.Attributes["team_name"])
			}
			next, err := res.Diff(t.Context(), after, cfg, meta)
			require.NoError(t, err)
			require.Equal(t, !rejected && !tc.rejectRename, next.Empty(), "%#v", next)
			if tc.rejectRename {
				require.NotContains(t, next.Attributes, "team_id")
			}
		})
	}
}
