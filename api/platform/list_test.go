package platform

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestIncompletePlatformList(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		next, body string
	}{
		{"missing range", 206, "", "[]"},
		{"repeated range", 206, "id repeated..", "[]"},
		{"missing list", 200, "", "null"},
		{"invalid JSON", 206, "id next..", "["},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if calls > 2 {
					t.Error("pagination did not stop")
					w.WriteHeader(500)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Next-Range", tc.next)
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer srv.Close()
			client := New(&config.Config{PlatformBaseURL: srv.URL})
			apps, err := client.ListApps(t.Context())
			require.Error(t, err)
			require.Nil(t, apps)
		})
	}
}

func TestInventoryCancellationDuringRateLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusTooManyRequests)
		cancel()
	}))
	defer srv.Close()
	client := New(&config.Config{PlatformBaseURL: srv.URL})
	start := time.Now()
	_, err := client.ListApps(ctx)
	require.ErrorIs(t, err, context.Canceled)
	require.Less(t, time.Since(start), 2*time.Second)
	require.Equal(t, 1, calls)
}

func TestConcurrentInventoryWithProxy(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("loopback inventory request sent to proxy: %s", r.URL.Path)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer proxy.Close()
	t.Setenv("HTTPS_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "localhost")
	t.Setenv("no_proxy", "")
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":"one"}]`)
	}))
	defer origin.Close()
	p := New(&config.Config{PlatformBaseURL: origin.URL})
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			apps, err := p.ListApps(t.Context())
			if err != nil || len(apps) != 1 {
				t.Errorf("apps: %v, %v", apps, err)
			}
		})
		wg.Go(func() {
			addons, err := p.ListAddons(t.Context(), "")
			if err != nil || len(addons) != 1 {
				t.Errorf("add-ons: %v, %v", addons, err)
			}
		})
	}
	wg.Wait()
}
