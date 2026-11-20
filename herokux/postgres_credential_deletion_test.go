package herokux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresCredentialDeletionChecksAttachments(t *testing.T) {
	const id = "db:report"
	const path = "/addons/db/addon-attachments"
	const credential = "/postgres/v0/databases/db/credentials/report"
	const other = `[{"app":{"name":"ordinary"},"namespace":null},{"app":{"name":"different"},"namespace":"credential:reporting"}]`
	for _, tc := range []struct {
		name      string
		status    int
		lastPage  string
		wantError string
	}{
		{"in use", 200, `[{"app":{"name":"worker"},"namespace":"credential:report"},{"app":{"name":"api"},"namespace_config":{"role":"report"}},{"app":{"name":"batch"},"namespace_config":{"cred":"report"}},{"app":{"name":"worker"},"namespace":"credential:report"}]`, "api, batch, worker"},
		{"unattached", 200, other, ""},
		{"lookup failed", 403, `{"id":"forbidden"}`, "403"},
		{"invalid attachment list", 200, `null`, "attachment"},
		{"incomplete pagination", 206, other, "range"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, r.Method+" "+r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == path:
					if r.Header.Get("Range") == "" {
						w.Header().Set("Next-Range", "id next..")
						w.WriteHeader(http.StatusPartialContent)
						_, _ = io.WriteString(w, other)
						return
					}
					assert.Equal(t, "id next..", r.Header.Get("Range"))
					w.WriteHeader(tc.status)
					_, _ = io.WriteString(w, tc.lastPage)
				case r.Method == http.MethodDelete && r.URL.Path == credential:
					if tc.wantError != "" {
						w.WriteHeader(409)
						_, _ = io.WriteString(w, `{"message":"unexpected delete"}`)
						return
					}
					_, _ = io.WriteString(w, "{}")
				case r.Method == http.MethodGet && r.URL.Path == credential:
					w.WriteHeader(404)
				default:
					t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
					w.WriteHeader(400)
				}
			}))
			defer srv.Close()
			client, err := api.New(config.PostgresBaseURL(srv.URL), config.PlatformBaseURL(srv.URL))
			require.NoError(t, err)
			r := resourceHerokuxPostgresCredential()
			d := r.Data(nil)
			d.SetId(id)
			ds := r.DeleteContext(t.Context(), d, &Config{API: client, PostgresCredentialDeleteVerifyTimeout: 1})
			reads := []string{"GET " + path, "GET " + path}
			if tc.wantError != "" {
				require.Equal(t, reads, calls, "an attached credential or a failed lookup must never reach DELETE")
				require.True(t, ds.HasError())
				require.Contains(t, ds[0].Summary, tc.wantError)
				require.Equal(t, id, d.Id())
				return
			}
			require.False(t, ds.HasError(), "%v", ds)
			require.Equal(t, append(reads, "DELETE "+credential, "GET "+credential), calls)
		})
	}
}
