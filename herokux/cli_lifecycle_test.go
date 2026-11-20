package herokux

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestPoolingAttachmentLifecycle(t *testing.T) {
	const app = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	status := 403
	posts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /client/v11/databases/my-db":
			fmt.Fprint(w, `{"info":[{"name":"Connection Pooling","values":["Available"]}]}`)
		case "POST /client/v11/databases/my-db/connection-pooling":
			posts++
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["app"] != app || body["credential"] != "default" {
				t.Errorf("unexpected attachment request %v", body)
			}
			fmt.Fprint(w, `{"id":"pool"}`)
		case "GET /addon-attachments/pool":
			w.WriteHeader(status)
			if status != 200 {
				fmt.Fprint(w, `{"message":"lookup failed"}`)
				return
			}
			fmt.Fprintf(w, `{"id":"pool","namespace":"connection-pooling:default","name":"DATABASE_CONNECTION_POOL","addon":{"id":"db-id","name":"my-db"},"app":{"id":%q}}`, app)
		case "GET /addons/db-id":
			fmt.Fprint(w, `{"id":"db-id","addon_service":{"name":"heroku-postgresql"}}`)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(400)
		}
	}))
	defer srv.Close()
	c := NewConfig()
	require.NoError(t, c.applySchema(schema.TestResourceDataRaw(t, New().Schema, map[string]any{"postgres_api_url": srv.URL, "platform_api_url": srv.URL})))
	require.NoError(t, c.initializeAPI())
	r := resourceHerokuxPostgresConnectionPooling()
	args := terraform.NewResourceConfigRaw(map[string]any{"app_id": app, "postgres_id": "my-db"})
	diff, err := r.Diff(t.Context(), nil, args, c)
	require.NoError(t, err)
	state, ds := r.Apply(t.Context(), nil, diff, c)
	require.True(t, ds.HasError())
	require.NotNil(t, state)
	require.Equal(t, "pool", state.ID)
	require.Equal(t, 1, posts)
	status = 200
	state, ds = r.RefreshWithoutUpgrade(t.Context(), state, c)
	require.False(t, ds.HasError(), "%v", ds)
	require.Equal(t, "my-db", state.Attributes["postgres_id"])
	diff, err = r.Diff(t.Context(), state, args, c)
	require.NoError(t, err)
	require.True(t, diff.Empty(), "%s", diff)
	status = 404
	state, ds = r.RefreshWithoutUpgrade(t.Context(), state, c)
	require.False(t, ds.HasError(), "%v", ds)
	require.Nil(t, state)
	d := r.Data(nil)
	d.SetId("pool")
	_, err = r.Importer.StateContext(t.Context(), d, c)
	require.Error(t, err)
}

func TestCLIResourceRefresh(t *testing.T) {
	for _, tc := range []struct {
		name            string
		r               *schema.Resource
		id, path        string
		status          int
		body            string
		missing, failed bool
	}{
		{"Kafka topic removed", resourceHerokuxKafkaTopic(), "cluster:orders", "/data/kafka/v0/clusters/cluster/topics", 200, `{"topics":[]}`, true, false},
		{"Kafka malformed topic list", resourceHerokuxKafkaTopic(), "cluster:orders", "/data/kafka/v0/clusters/cluster/topics", 200, `{}`, false, true},
		{"Kafka missing cluster", resourceHerokuxKafkaTopic(), "cluster:orders", "/data/kafka/v0/clusters/cluster/topics", 404, `{}`, false, true},
		{"data link removed", resourceHerokuxPostgresDataLink(), "db:one", "/client/v11/databases/db/links", 200, `[]`, true, false},
		{"data link invalid list", resourceHerokuxPostgresDataLink(), "db:one", "/client/v11/databases/db/links", 200, `null`, false, true},
		{"connector removed", resourceHerokuxDataConnector(), "one", "/data/cdc/v0/connectors/one", 404, `{}`, true, false},
		{"connector malformed", resourceHerokuxDataConnector(), "one", "/data/cdc/v0/connectors/one", 200, `{}`, false, true},
		{"credential removed", resourceHerokuxPostgresCredential(), "db:one", "/postgres/v0/databases/db/credentials/one", 404, `{}`, true, false},
		{"credential malformed", resourceHerokuxPostgresCredential(), "db:one", "/postgres/v0/databases/db/credentials/one", 200, `null`, false, true},
		{"mTLS malformed", resourceHerokuxPostgresMTLS(), "db", "/postgres/v0/databases/db/tls-endpoint", 200, `{}`, false, true},
		{"mTLS certificate malformed", resourceHerokuxPostgresMTLSCertificate(), "db:one", "/postgres/v0/databases/db/tls-endpoint/certificates/one", 200, `{}`, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != tc.path {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer srv.Close()
			c, err := api.New(config.PostgresBaseURL(srv.URL), config.KafkaBaseURL(srv.URL))
			require.NoError(t, err)
			d := tc.r.Data(nil)
			d.SetId(tc.id)
			ds := tc.r.ReadContext(t.Context(), d, &Config{API: c})
			require.Equal(t, tc.failed, ds.HasError(), "%v", ds)
			if tc.missing {
				require.Empty(t, d.Id())
			} else {
				require.Equal(t, tc.id, d.Id())
			}
		})
	}
}

func TestResourceTransportErrorsKeepState(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	c, err := api.New(config.PostgresBaseURL(srv.URL), config.KafkaBaseURL(srv.URL))
	require.NoError(t, err)
	r := resourceHerokuxPostgresMTLS()
	d := r.Data(nil)
	d.SetId("db")
	require.True(t, r.ReadContext(t.Context(), d, &Config{API: c}).HasError())
	require.Equal(t, "db", d.Id())
	for _, refresh := range []func() (interface{}, string, error){
		MTLSDeletionStateRefreshFunc(c, "db"), MTLSCertificateDeletionStateRefreshFunc(c, "db", "cert"),
		postgresCredentialDeletionStateRefreshFunc(c, "db", "one"), DataConnectorDeleteStateRefreshFunc(c, "one"),
		PrivatelinkDeleteStateRefreshFunc(c, "db"), topicCreationStateRefreshFunc(c, "cluster", "one", 1),
	} {
		_, _, err := refresh()
		require.Error(t, err)
	}
}

func TestRegistryImageSchemas(t *testing.T) {
	hex := strings.Repeat("a", 64)
	for _, tc := range []struct {
		body   string
		layers int
		valid  bool
	}{
		{fmt.Sprintf(`{"schemaVersion":1,"history":[{"v1Compatibility":%q}],"fsLayers":[{},{}]}`, "{\"id\":\""+hex+"\"}"), 2, true},
		{fmt.Sprintf(`{"schemaVersion":2,"config":{"digest":"sha256:%s","size":42},"layers":[{}]}`, hex), 1, true},
		{`{"schemaVersion":1,"history":[]}`, 0, false}, {`{"schemaVersion":2,"config":{}}`, 0, false}, {`null`, 0, false},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
			fmt.Fprint(w, tc.body)
		}))
		client, err := api.New(config.RegistryBaseURL(srv.URL))
		require.NoError(t, err)
		r := dataSourceHerokuxRegistryImage()
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"app_id": "6fae1ee0-c034-4775-a798-890bc64f98eb", "process_type": "web"})
		ds := r.ReadContext(t.Context(), d, &Config{API: client})
		srv.Close()
		require.Equal(t, tc.valid, !ds.HasError(), "%v", ds)
		if tc.valid {
			require.Equal(t, "sha256:"+hex, d.Get("digest"))
			require.Equal(t, tc.layers, d.Get("number_of_layers"))
		} else {
			require.Empty(t, d.Id())
		}
	}
}

func TestContainerReleaseWaitsForNewMatchingRelease(t *testing.T) {
	const desc = "Deployed web (aaaaaaaaaaaa)"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/apps/app/releases":
			fmt.Fprintf(w, `[{"version":12,"description":"unrelated"},{"id":"new","version":11,"description":%q,"status":"pending"},{"id":"old","version":10,"description":%q,"status":"succeeded"}]`, desc, desc)
		case "/apps/app/releases/new":
			fmt.Fprint(w, `{"id":"new","version":11,"status":"failed"}`)
		default:
			t.Error(r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	c := heroku.NewService(srv.Client())
	c.URL = srv.URL
	poll := containerReleaseStateRefreshFunc(t.Context(), c, "app", "sha256:"+strings.Repeat("a", 64), "web", 10)
	_, status, err := poll()
	require.NoError(t, err)
	require.Equal(t, "pending", status)
	_, status, err = poll()
	require.NoError(t, err)
	require.Equal(t, "failed", status)
}

func TestSensitiveProviderAndSpaceFields(t *testing.T) {
	require.True(t, New().Schema["api_key"].Sensitive)
	require.True(t, New().Schema["headers"].Sensitive)
	require.True(t, resourceHerokuxShieldPrivateSpace().Schema["log_drain"].Sensitive)
	const secret = "test-only-credential"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, secret)
	}))
	defer srv.Close()
	c, err := api.New(config.PlatformBaseURL(srv.URL))
	require.NoError(t, err)
	_, _, err = c.Platform.SetSpaceLogDrain("space", "https://logs.example/"+secret)
	require.ErrorContains(t, err, "422")
	require.NotContains(t, err.Error(), secret)
}

func TestPoolingWaitSurvivesRepeatedNotFound(t *testing.T) {
	lookups := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/addon-attachments/pool":
			lookups++
			if lookups <= 21 {
				w.WriteHeader(404)
				fmt.Fprint(w, `{"id":"not_found","message":"attachment not visible yet"}`)
				return
			}
			fmt.Fprint(w, `{"id":"pool","namespace":"connection-pooling:default","name":"POOL","addon":{"id":"db"},"app":{"id":"app"}}`)
		case "/addons/db":
			fmt.Fprint(w, `{"addon_service":{"name":"heroku-postgresql"}}`)
		default:
			t.Error(r.URL.Path)
			w.WriteHeader(400)
		}
	}))
	defer srv.Close()
	c := heroku.NewService(&http.Client{Transport: &heroku.Transport{Transport: srv.Client().Transport}})
	c.URL = srv.URL
	wait := poolingAttachmentWait(t.Context(), c, "pool")
	wait.PollInterval = time.Millisecond
	result, err := wait.WaitForStateContext(t.Context())
	require.NoError(t, err)
	require.Equal(t, "pool", result.(*heroku.AddOnAttachment).ID)
	require.Equal(t, 22, lookups)
}

func TestContainerReleaseUnchangedVersion(t *testing.T) {
	const desc = "Deployed web (aaaaaaaaaaaa)"
	for _, tc := range []struct{ name, body, status string }{
		{"pending", `[{"id":"same","version":10,"description":"` + desc + `","status":"pending"}]`, "pending"},
		{"failed", `[{"id":"same","version":10,"description":"` + desc + `","status":"failed"}]`, "failed"},
		{"unrelated latest", `[{"version":10,"description":"Scaled worker","status":"succeeded"},{"id":"same","version":9,"description":"` + desc + `","status":"pending"}]`, "pending"},
		{"unrelated only", `[{"version":10,"description":"Scaled worker","status":"succeeded"}]`, "succeeded"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/apps/app/releases":
					fmt.Fprint(w, tc.body)
				case "/apps/app/releases/same":
					fmt.Fprint(w, `{"id":"same","status":"failed"}`)
				default:
					t.Error(r.URL.Path)
					w.WriteHeader(400)
				}
			}))
			defer srv.Close()
			c := heroku.NewService(srv.Client())
			c.URL = srv.URL
			poll := containerReleaseStateRefreshFunc(t.Context(), c, "app", "sha256:"+strings.Repeat("a", 64), "web", 10)
			_, status, err := poll()
			require.NoError(t, err)
			require.Equal(t, tc.status, status)
			if status == "pending" {
				_, status, err = poll()
				require.NoError(t, err)
				require.Equal(t, "failed", status)
			}
		})
	}
}

func TestPipelineMemberUUIDCasing(t *testing.T) {
	const pipeline = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	const email = "developer@example.org"
	upper := strings.ToUpper(pipeline)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.EqualFold(r.URL.Path, "/pipelines/"+pipeline+"/ephemeral-app-collaborators") {
			t.Error(r.URL.Path)
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"id":"member","pipeline":{"id":%q},"user":{"email":%q},"permissions":[{"name":"view"}]}]`, pipeline, email)
	}))
	defer srv.Close()
	c, err := api.New(config.PlatformBaseURL(srv.URL))
	require.NoError(t, err)
	meta := &Config{API: c}
	r := resourceHerokuxPipelineMember()
	d := r.Data(nil)
	d.SetId(upper + ":" + email)
	entries, err := r.Importer.StateContext(t.Context(), d, meta)
	require.NoError(t, err)
	state, ds := r.RefreshWithoutUpgrade(t.Context(), entries[0].State(), meta)
	require.False(t, ds.HasError(), "%v", ds)
	require.Equal(t, upper, state.Attributes["pipeline_id"])
	for _, id := range []string{upper, pipeline} {
		cfg := terraform.NewResourceConfigRaw(map[string]any{"pipeline_id": id, "email": email, "permissions": []any{"view"}})
		diff, err := r.Diff(t.Context(), state, cfg, meta)
		require.NoError(t, err)
		require.True(t, diff.Empty(), "%v", diff)
	}
}
