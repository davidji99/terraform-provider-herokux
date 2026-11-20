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

func TestKafkaConsumerGroupRefreshAfterDeletion(t *testing.T) {
	const id = "f8a8fa1a-bbe4-4ce8-a06c-b485ae94d7f0:orders"
	args := map[string]any{"kafka_id": "f8a8fa1a-bbe4-4ce8-a06c-b485ae94d7f0", "name": "orders"}
	r := resourceHerokuxKafkaConsumerGroup()
	for _, tc := range []struct {
		name       string
		status     int
		body, want string
	}{
		{"present", 200, `{"consumer_groups":[{"name":"orders"}]}`, "present"},
		{"empty", 200, `{"consumer_groups":[]}`, "gone"},
		{"different name", 200, `{"consumer_groups":[{"name":"orders-archive"}]}`, "gone"},
		{"missing list", 200, `{}`, "failed"},
		{"missing name", 200, `{"consumer_groups":[{}]}`, "failed"},
		{"forbidden", 403, `{"message":"permission denied"}`, "failed"},
		{"service unavailable", 503, `{"message":"unavailable"}`, "failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodGet || req.URL.Path != "/data/kafka/v0/clusters/f8a8fa1a-bbe4-4ce8-a06c-b485ae94d7f0/consumer_groups" {
					t.Errorf("unexpected %s %s", req.Method, req.URL.Path)
					w.WriteHeader(400)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer srv.Close()
			client, err := api.New(config.KafkaBaseURL(srv.URL))
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
				require.Equal(t, "orders", refreshed.Attributes["name"])
			case "gone":
				require.False(t, ds.HasError(), "%v", ds)
				require.Nil(t, refreshed)
				plan, err := r.Diff(t.Context(), refreshed, terraform.NewResourceConfigRaw(args), meta)
				require.NoError(t, err)
				require.False(t, plan.Empty(), "declared resource should be created again")
				if tc.name == "empty" {
					d.SetId("f8a8fa1a-bbe4-4ce8-a06c-b485ae94d7f0:orders")
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
