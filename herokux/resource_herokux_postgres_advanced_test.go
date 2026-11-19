package herokux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/davidji99/terraform-provider-herokux/api/postgres/advanced"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const advancedTestDB = "29138cab-acf1-40c6-ac83-56697c8f0221"
const advancedTestPath = "/data/postgres/v1/" + advancedTestDB

type advancedRequest struct {
	method, path, body string
	status             int
	response           string
}

type advancedTestResource struct {
	t        *testing.T
	requests chan advancedRequest
	resource *schema.Resource
	meta     *Config
	state    *terraform.InstanceState
}

func newAdvancedTestResource(t *testing.T, resourceName string) *advancedTestResource {
	t.Helper()
	requests := make(chan advancedRequest, 32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case expected := <-requests:
			assert.Equal(t, expected.method, r.Method)
			assert.Equal(t, advancedTestPath+expected.path, r.URL.EscapedPath())
			assert.Equal(t, "Bearer local-token", r.Header.Get("Authorization"))
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			if expected.body == "" {
				assert.Empty(t, body)
			} else {
				assert.JSONEq(t, expected.body, string(body))
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(expected.status)
			_, _ = io.WriteString(w, expected.response)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(func() {
		server.Close()
		assert.Empty(t, requests, "unconsumed API requests")
	})
	client, err := api.New(config.PostgresAdvancedBaseURL(server.URL), config.APIToken("local-token"))
	require.NoError(t, err)
	return &advancedTestResource{t: t, requests: requests, resource: New().ResourcesMap[resourceName], meta: &Config{API: client}}
}

func (f *advancedTestResource) expect(method, path, body string, status int, response string) {
	f.requests <- advancedRequest{method, path, body, status, response}
}

func (f *advancedTestResource) plan(args map[string]any) *terraform.InstanceDiff {
	f.t.Helper()
	diff, err := f.resource.Diff(f.t.Context(), f.state, terraform.NewResourceConfigRaw(args), f.meta)
	require.NoError(f.t, err)
	return diff
}

func (f *advancedTestResource) apply(args map[string]any) {
	f.t.Helper()
	next, ds := f.resource.Apply(f.t.Context(), f.state, f.plan(args), f.meta)
	require.False(f.t, ds.HasError(), "%v", ds)
	f.state = next
}

func (f *advancedTestResource) refresh() {
	f.t.Helper()
	next, ds := f.resource.RefreshWithoutUpgrade(f.t.Context(), f.state, f.meta)
	require.False(f.t, ds.HasError(), "%v", ds)
	f.state = next
}

func (f *advancedTestResource) importID(id string) {
	f.t.Helper()
	d := schema.TestResourceDataRaw(f.t, f.resource.Schema, map[string]any{})
	d.SetId(id)
	result, err := f.resource.Importer.StateContext(f.t.Context(), d, f.meta)
	require.NoError(f.t, err)
	f.state = result[0].State()
	f.refresh()
}

func (f *advancedTestResource) destroy() {
	f.t.Helper()
	next, ds := f.resource.Apply(f.t.Context(), f.state, &terraform.InstanceDiff{Destroy: true}, f.meta)
	require.False(f.t, ds.HasError(), "%v", ds)
	require.Nil(f.t, next)
	f.state = next
}

func advancedAcceptanceDB(t *testing.T) string {
	t.Helper()
	testAccConfig.SkipUnlessAccTest(t)
	db := os.Getenv("HEROKUX_POSTGRES_ADVANCED_ID")
	if db == "" {
		t.Skip("set HEROKUX_POSTGRES_ADVANCED_ID to a dedicated Postgres Advanced test database")
	}
	testAccPreCheck(t)
	return db
}

func advancedAcceptanceClient(t *testing.T) *advanced.Client {
	t.Helper()
	opts := []config.Option{config.APIToken(os.Getenv("HEROKU_API_KEY"))}
	if endpoint := os.Getenv("HEROKUX_POSTGRES_ADVANCED_API_URL"); endpoint != "" {
		opts = append(opts, config.PostgresAdvancedBaseURL(endpoint))
	}
	client, err := api.New(opts...)
	require.NoError(t, err)
	return client.PostgresAdvanced
}
