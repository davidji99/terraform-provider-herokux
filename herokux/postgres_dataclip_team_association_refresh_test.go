package herokux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestPostgresDataclipTeamAssociationRefreshAfterDeletion(t *testing.T) {
	const id = "share-id"
	args := map[string]any{"dataclip_id": "6fae1ee0-c034-4775-a798-890bc64f98eb", "dataclip_slug": "clip", "team_id": "38b6e411-1361-4adf-ace6-734c9a95513d"}
	r := resourceHerokuxPostgresDataclipTeamAssociation()
	for _, tc := range []struct {
		name       string
		status     int
		body, want string
	}{
		{"present", 200, `{"data":{"clip":{"id":"6fae1ee0-c034-4775-a798-890bc64f98eb","slug":"clip","team_shares":[{"id":"share-id","clip_id":"6fae1ee0-c034-4775-a798-890bc64f98eb","shared_by":{"email":"owner@example.test"},"shared_with":{"id":"38b6e411-1361-4adf-ace6-734c9a95513d","name":"finance"}}]}}}`, "present"},
		{"empty", 200, `{"data":{"clip":{"id":"6fae1ee0-c034-4775-a798-890bc64f98eb","slug":"clip","team_shares":[]}}}`, "gone"},
		{"new share ID for same recipient", 200, `{"data":{"clip":{"id":"6fae1ee0-c034-4775-a798-890bc64f98eb","slug":"clip","team_shares":[{"id":"replacement","clip_id":"6fae1ee0-c034-4775-a798-890bc64f98eb","shared_by":{"email":"owner@example.test"},"shared_with":{"id":"38b6e411-1361-4adf-ace6-734c9a95513d","name":"finance"}}]}}}`, "gone"},
		{"missing list", 200, `{"data":{"clip":{"id":"6fae1ee0-c034-4775-a798-890bc64f98eb","slug":"clip","team_shares":null}}}`, "failed"},
		{"missing share ID", 200, `{"data":{"clip":{"id":"6fae1ee0-c034-4775-a798-890bc64f98eb","slug":"clip","team_shares":[{}]}}}`, "failed"},
		{"partial GraphQL result", 200, `{"data":{"clip":{"id":"6fae1ee0-c034-4775-a798-890bc64f98eb","slug":"clip","team_shares":[]}},"errors":[{"message":"permission denied"}]}`, "failed"},
		{"unavailable clip", 200, `{"data":{"clip":null}}`, "failed"},
		{"forbidden", 403, `{"message":"permission denied"}`, "failed"},
		{"service unavailable", 503, `{"message":"unavailable"}`, "failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodGet || req.URL.Path != "/graphql" {
					t.Errorf("unexpected %s %s", req.Method, req.URL.Path)
					w.WriteHeader(400)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer srv.Close()
			client, err := api.New(config.DataBaseURL(srv.URL))
			require.NoError(t, err)
			meta := &Config{API: client}
			d := schema.TestResourceDataRaw(t, r.Schema, args)
			d.SetId(id)
			refreshed, ds := r.RefreshWithoutUpgrade(t.Context(), d.State(), meta)
			switch tc.want {
			case "present":
				require.False(t, ds.HasError(), "%v", ds)
				require.NotNil(t, refreshed)
				require.Equal(t, id, refreshed.ID)
				require.Equal(t, "finance", refreshed.Attributes["team_name"])
			case "gone":
				require.False(t, ds.HasError(), "%v", ds)
				require.Nil(t, refreshed)
				plan, err := r.Diff(t.Context(), refreshed, terraform.NewResourceConfigRaw(args), meta)
				require.NoError(t, err)
				require.False(t, plan.Empty(), "declared resource should be created again")
				if tc.name == "empty" {
					d.SetId("clip:finance")
					_, err := r.Importer.StateContext(t.Context(), d, meta)
					require.Error(t, err)
				}
			case "failed":
				require.True(t, ds.HasError(), "an inconclusive response must keep state")
				require.NotNil(t, refreshed)
				require.Equal(t, id, refreshed.ID)
			}
		})
	}
}
