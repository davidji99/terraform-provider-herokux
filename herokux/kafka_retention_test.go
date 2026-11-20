package herokux

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestKafkaTopicRetentionLimits(t *testing.T) {
	for _, tc := range []struct {
		minimum int
		value   string
		valid   bool
	}{
		{21600000, "6h", true}, {86400000, "6h", false}, {21600000, "5h", false}, {86400000, "1d", true}, {86400000, "disable", true},
	} {
		t.Run(fmt.Sprintf("%d_%s", tc.minimum, tc.value), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/data/kafka/v0/clusters/cluster" {
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"limits":{"minimum_retention_ms":%d}}`, tc.minimum)
			}))
			defer server.Close()
			c, err := api.New(config.KafkaBaseURL(server.URL))
			require.NoError(t, err)
			_, errs := validateRetentionTime(tc.value, "retention_time")
			ms, err := topicRetentionMS(c.Kafka, "cluster", tc.value)
			require.Equal(t, tc.valid, len(errs) == 0 && err == nil)
			if tc.value == "disable" {
				require.Nil(t, ms)
			}
		})
	}
	for _, bad := range []string{"6hgarbage", "xdisable", "999999999999999999w"} {
		_, errs := validateRetentionTime(bad, "retention_time")
		require.NotEmpty(t, errs)
	}
	require.Equal(t, "1d", resourceHerokuxKafkaTopic().Schema["retention_time"].Default)
}

func TestKafkaEquivalentRetentionPlan(t *testing.T) {
	r := resourceHerokuxKafkaTopic()
	args := map[string]any{"kafka_id": "6fae1ee0-c034-4775-a798-890bc64f98eb", "name": "orders", "partitions": 3, "retention_time": "24h"}
	d := schema.TestResourceDataRaw(t, r.Schema, args)
	d.SetId(args["kafka_id"].(string) + ":orders")
	require.NoError(t, d.Set("retention_time", "1d"))
	for _, tc := range []struct {
		value    string
		wantDiff bool
	}{{"24h", false}, {"1440m", false}, {"2d", true}, {"disable", true}} {
		args["retention_time"] = tc.value
		plan, err := r.Diff(t.Context(), d.State(), terraform.NewResourceConfigRaw(args), nil)
		require.NoError(t, err)
		require.Equal(t, tc.wantDiff, !plan.Empty(), "%s: %v", tc.value, plan)
	}
}
