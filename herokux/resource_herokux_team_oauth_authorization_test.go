package herokux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const teamOAuthID = "ac19340e-e020-417c-8ab6-4646796eb156"
const teamOAuthPath = "/teams/example/oauth/authorizations"
const oauthDetailPath = "/oauth/authorizations/" + teamOAuthID
const authFields = `"id":"` + teamOAuthID + `","scope":["read"],"description":"reports"`
const tokenID = "e5acd738-cac6-4987-8239-a6d7a523c21e"
const authMetadata = `{` + authFields + `,"access_token":{"id":"` + tokenID + `","expires_in":75}}`

type teamOAuthStep struct {
	method, path, body               string
	status                           int
	response, rangeHeader, nextRange string
}

func teamOAuthFixture(t *testing.T, steps ...teamOAuthStep) (*schema.Resource, *Config) {
	t.Helper()
	queue := make(chan teamOAuthStep, len(steps))
	for _, step := range steps {
		queue <- step
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var step teamOAuthStep
		select {
		case step = <-queue:
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		assert.Equal(t, step.method, r.Method)
		assert.Equal(t, step.path, r.URL.EscapedPath())
		if step.path == teamOAuthPath {
			assert.Equal(t, "application/vnd.heroku+json; version=3.sdk", r.Header.Get("Accept"))
			assert.Equal(t, "Bearer local-token", r.Header.Get("Authorization"))
		}
		assert.Equal(t, step.rangeHeader, r.Header.Get("Range"))
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		if step.body == "" {
			assert.Empty(t, body)
		} else {
			assert.JSONEq(t, step.body, string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		if step.nextRange != "" {
			w.Header().Set("Next-Range", step.nextRange)
		}
		w.WriteHeader(step.status)
		if step.response != "" {
			_, _ = io.WriteString(w, step.response)
		}
	}))
	t.Cleanup(func() {
		server.Close()
		assert.Empty(t, queue)
	})
	p := New()
	c := NewConfig()
	require.NoError(t, c.applySchema(schema.TestResourceDataRaw(t, p.Schema, map[string]any{"platform_api_url": server.URL})))
	c.token = "local-token"
	require.NoError(t, c.initializeAPI())
	r := p.ResourcesMap["herokux_team_oauth_authorization"]
	require.NotNil(t, r)
	return r, c
}

func TestTeamOAuthAuthorization(t *testing.T) {
	const list = "[" + authMetadata + "]"
	renamed := strings.ReplaceAll(authMetadata, "reports", "monthly reports")
	const created = `{` + authFields + `,"access_token":{"id":"` + tokenID + `","expires_in":100,"token":"creation-only"}}`
	r, c := teamOAuthFixture(t,
		teamOAuthStep{method: "POST", path: teamOAuthPath, status: 201, body: `{"description":"reports","expires_in":100,"scope":["read"]}`, response: created},
		teamOAuthStep{method: "GET", path: teamOAuthPath, status: 200, response: `[{` + authFields + `}]`},
		teamOAuthStep{method: "GET", path: teamOAuthPath, status: 200, response: list},
		teamOAuthStep{method: "PATCH", path: oauthDetailPath, status: 200, body: `{"description":"monthly reports"}`, response: renamed},
		teamOAuthStep{method: "GET", path: teamOAuthPath, status: 200, response: "[" + renamed + "]"},
		teamOAuthStep{method: "GET", path: teamOAuthPath, status: 200, response: list},
		teamOAuthStep{method: "DELETE", path: oauthDetailPath, status: 204},
	)
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"team": "example", "scope": []any{"read"}, "description": "reports", "time_to_live": 100})
	require.Empty(t, r.CreateContext(t.Context(), d, c))
	require.Equal(t, teamOAuthID, d.Id())
	require.True(t, r.Schema["access_token"].Sensitive)
	require.Equal(t, "creation-only", d.Get("access_token"))
	require.Empty(t, r.ReadContext(t.Context(), d, c))
	require.Equal(t, "creation-only", d.Get("access_token"))
	require.NoError(t, d.Set("description", "monthly reports"))
	require.Empty(t, r.UpdateContext(t.Context(), d, c))
	require.Equal(t, "monthly reports", d.Get("description"))
	require.Equal(t, "creation-only", d.Get("access_token"))
	require.Equal(t, 75, d.Get("expires_in"))
	require.Equal(t, tokenID, d.Get("token_id"))
	require.Empty(t, r.DeleteContext(t.Context(), d, c))
	require.Empty(t, d.Id())
}

func TestTeamOAuthImport(t *testing.T) {
	// Even an unexpected secret in a list response must never enter imported state.
	const list = `[{` + authFields + `,"access_token":{"id":"` + tokenID + `","expires_in":75,"token":"ignore-this"}}]`
	r, c := teamOAuthFixture(t,
		teamOAuthStep{method: "GET", path: teamOAuthPath, status: 206, response: "[]", nextRange: "id next.."},
		teamOAuthStep{method: "GET", path: teamOAuthPath, status: 200, rangeHeader: "id next..", response: strings.Replace(list, teamOAuthID, strings.ToUpper(teamOAuthID), 1)},
	)
	d := r.Data(nil)
	for _, invalid := range []string{teamOAuthID, "/" + teamOAuthID, "../" + teamOAuthID, "example/not-a-uuid", "example/" + teamOAuthID + "/extra"} {
		d.SetId(invalid)
		_, err := r.Importer.StateContext(t.Context(), d, c)
		require.ErrorContains(t, err, "<team>/<authorization-uuid>")
	}
	d.SetId("example/" + strings.ToUpper(teamOAuthID))
	_, err := r.Importer.StateContext(t.Context(), d, c)
	require.NoError(t, err)
	require.Empty(t, r.ReadContext(t.Context(), d, c))
	require.Equal(t, teamOAuthID, d.Id())
	state := d.State().Attributes
	require.Equal(t, "example", state["team"])
	require.Equal(t, "reports", state["description"])
	require.Equal(t, tokenID, state["token_id"])
	require.Equal(t, "75", state["expires_in"])
	require.NotContains(t, state, "access_token")
	cfg := terraform.NewResourceConfigRaw(map[string]any{"team": "example", "scope": []any{"read"}, "description": "reports"})
	diff, err := r.Diff(t.Context(), d.State(), cfg, c)
	require.NoError(t, err)
	require.True(t, diff.Empty(), "import should produce an empty follow-up plan")

}

func TestTeamOAuthOwnershipAndErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		list   string
	}{
		{"denied", 403, `{"message":"secret-marker"}`}, {"team missing", 404, `{}`},
		{"server failure", 500, `{"message":"secret-marker"}`}, {"incomplete page", 206, "[]"},
		{"malformed list", 200, "null"}, {"missing identity", 200, `[{}]`},
		{"web grant", 200, `[{` + authFields + `,"client":{"id":"client"}}]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			step := teamOAuthStep{method: "GET", path: teamOAuthPath, status: tc.status, response: tc.list}
			r, c := teamOAuthFixture(t, step, step, step)
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"team": "example", "scope": []any{"read"}})
			d.SetId(teamOAuthID)
			for _, action := range []schema.ReadContextFunc{r.ReadContext, schema.ReadContextFunc(r.UpdateContext), schema.ReadContextFunc(r.DeleteContext)} {
				ds := action(t.Context(), d, c)
				require.True(t, ds.HasError())
				require.NotContains(t, ds[0].Summary+ds[0].Detail, "secret-marker")
				require.Equal(t, teamOAuthID, d.Id())
			}
		})
	}
	t.Run("absent from team", func(t *testing.T) {
		step := teamOAuthStep{method: "GET", path: teamOAuthPath, status: 200, response: "[]"}
		r, c := teamOAuthFixture(t, step, step, step)
		d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"team": "example", "scope": []any{"read"}})
		d.SetId(teamOAuthID)
		require.Empty(t, r.ReadContext(t.Context(), d, c))
		require.Empty(t, d.Id())
		d.SetId(teamOAuthID)
		require.True(t, r.UpdateContext(t.Context(), d, c).HasError())
		d.SetId(teamOAuthID)
		require.Empty(t, r.DeleteContext(t.Context(), d, c))
		require.Empty(t, d.Id())
	})
}

func TestTeamOAuthCreateWithoutToken(t *testing.T) {
	r, c := teamOAuthFixture(t, teamOAuthStep{
		method: "POST", path: teamOAuthPath, status: 201, body: `{"scope":["read"]}`, response: authMetadata,
	})
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"team": "example", "scope": []any{"read"}})
	ds := r.CreateContext(t.Context(), d, c)
	require.True(t, ds.HasError())
	require.Equal(t, teamOAuthID, d.Id())
	require.NotContains(t, d.State().Attributes, "access_token")
}
