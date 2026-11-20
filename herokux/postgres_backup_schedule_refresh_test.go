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

func TestPostgresBackupScheduleRefreshAfterDeletion(t *testing.T) {
	const id = "schedule-id"
	args := map[string]any{"postgres_id": "db", "hour": 2, "timezone": "UTC"}
	r := resourceHerokuxPostgresBackupSchedule()
	for _, tc := range []struct {
		name       string
		status     int
		body, want string
	}{
		{"present", 200, `[{"uuid":"schedule-id","hour":2,"timezone":"UTC","name":"daily"}]`, "present"},
		{"empty", 200, `[]`, "gone"},
		{"different ID", 200, `[{"uuid":"replacement","hour":8,"timezone":"UTC"}]`, "gone"},
		{"missing list", 200, `null`, "failed"},
		{"missing ID", 200, `[{}]`, "failed"},
		{"forbidden", 403, `{"message":"permission denied"}`, "failed"},
		{"service unavailable", 503, `{"message":"unavailable"}`, "failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodGet || req.URL.Path != "/client/v11/databases/db/transfer-schedules" {
					t.Errorf("unexpected %s %s", req.Method, req.URL.Path)
					w.WriteHeader(400)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer srv.Close()
			client, err := api.New(config.PostgresBaseURL(srv.URL))
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
				require.Equal(t, "daily", refreshed.Attributes["name"])
			case "gone":
				require.False(t, ds.HasError(), "%v", ds)
				require.Nil(t, refreshed)
				plan, err := r.Diff(t.Context(), refreshed, terraform.NewResourceConfigRaw(args), meta)
				require.NoError(t, err)
				require.False(t, plan.Empty(), "declared resource should be created again")
				if tc.name == "empty" {
					d.SetId("db")
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
