package platform

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConcurrentPlatformVariantHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/spaces/space/log-drain":
			assert.Equal(t, DogwoodAcceptHeader, r.Header.Get("Accept"))
			fmt.Fprint(w, "{}")
		case "/pipelines/pipeline/ephemeral-app-collaborators":
			assert.Equal(t, PipelineCollaboratorsAcceptHeader, r.Header.Get("Accept"))
			fmt.Fprint(w, "[]")
		case "/addons/db/addon-attachments":
			assert.Equal(t, DefaultAcceptHeader, r.Header.Get("Accept"))
			fmt.Fprint(w, "[]")
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	client := New(&config.Config{PlatformBaseURL: server.URL})
	var workers sync.WaitGroup
	for _, request := range []func() error{
		func() error { _, _, err := client.GetSpaceLogDrain("space"); return err },
		func() error { _, _, err := client.ListPipelineMembers("pipeline"); return err },
		func() error { _, err := client.PostgresCredentialApps(t.Context(), "db", "report"); return err },
	} {
		workers.Go(func() {
			for range 10 {
				if err := request(); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	workers.Wait()
}
