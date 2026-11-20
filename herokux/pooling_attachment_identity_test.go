package herokux

import (
	"encoding/json"
	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostgresPoolingAttachmentIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, namespace, service, alias string
		valid                           bool
	}{
		{"custom", "connection-pooling:default", "heroku-postgresql", "CUSTOM_POOL", true},
		{"misleading name", "", "heroku-postgresql", "DATABASE_CONNECTION_POOL", false},
		{"credential", "credential:report", "heroku-postgresql", "REPORT", false},
		{"other service", "connection-pooling:default", "heroku-redis", "CUSTOM", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deletes := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/addon-attachments/attachment":
					if r.Method == http.MethodDelete {
						deletes++
						w.WriteHeader(http.StatusNoContent)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"id": "attachment", "name": tc.alias, "namespace": tc.namespace, "addon": map[string]string{"id": "db"}, "app": map[string]string{"id": "app"}})
				case "/addons/db":
					_ = json.NewEncoder(w).Encode(map[string]any{"id": "db", "addon_service": map[string]string{"name": tc.service}})
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(404)
				}
			}))
			defer srv.Close()
			client := heroku.NewService(srv.Client())
			client.URL = srv.URL
			r := resourceHerokuxPostgresConnectionPooling()
			meta := &Config{PlatformAPI: client}
			d := r.Data(nil)
			d.SetId("attachment")
			_, err := r.Importer.StateContext(t.Context(), d, meta)
			require.Equal(t, tc.valid, err == nil, "%v", err)
			ds := r.ReadContext(t.Context(), d, meta)
			require.Equal(t, !tc.valid, ds.HasError())
			ds = r.DeleteContext(t.Context(), d, meta)
			if tc.valid {
				require.False(t, ds.HasError(), "%v", ds)
				require.Equal(t, 1, deletes)
			} else {
				require.True(t, ds.HasError())
				require.Zero(t, deletes)
				require.Equal(t, "attachment", d.Id())
			}
		})
	}
}
