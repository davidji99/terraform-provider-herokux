package herokux

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/davidji99/terraform-provider-herokux/api/postgres"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrivatelinkAccountSwaps(t *testing.T) {
	const addon = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	const path = "/private-link/v0/databases/" + addon
	const old = "100000000001"
	const keep = "100000000002"
	const first = "100000000003"
	const second = "100000000004"

	interval := StateRefreshPollInterval
	StateRefreshPollInterval = time.Millisecond
	t.Cleanup(func() { StateRefreshPollInterval = interval })

	for _, scenario := range []string{"swap", "addition rejected", "activation lookup fails then retries", "pure revocation", "TerraformCLI no-refresh retry"} {
		t.Run(scenario, func(t *testing.T) {
			// An unrelated account can be provisioning without blocking a revocation.
			accounts := map[string]string{old: "Active", keep: "Provisioning"}
			var calls []string
			var poll int
			additionsSubmitted := false
			rejectRead := scenario == "activation lookup fails then retries" || scenario == "TerraformCLI no-refresh retry"
			advance := false
			var mu sync.Mutex
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				calls = append(calls, r.Method)
				w.Header().Set("Content-Type", "application/json")
				switch r.Method {
				case http.MethodPut, http.MethodPatch:
					assert.Equal(t, path+"/allowed_accounts", r.URL.Path)
					var body struct {
						Accounts []string `json:"allowed_accounts"`
					}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
						w.WriteHeader(400)
						return
					}
					if r.Method == http.MethodPut {
						assert.ElementsMatch(t, []string{first, second}, body.Accounts)
						if scenario == "addition rejected" {
							w.WriteHeader(422)
							return
						}
						additionsSubmitted = true
						accounts[first] = "Active"
						accounts[second] = "Provisioning"
					} else {
						assert.Equal(t, []string{old}, body.Accounts)
						if scenario != "pure revocation" {
							assert.Equal(t, "Active", accounts[first])
							assert.Equal(t, "Active", accounts[second])
						}
						delete(accounts, old)
					}
					w.WriteHeader(http.StatusAccepted)
					_, _ = w.Write([]byte("{}"))
				case http.MethodGet:
					assert.Equal(t, path, r.URL.Path)
					if additionsSubmitted && rejectRead {
						rejectRead = false
						w.WriteHeader(403)
						return
					}
					if additionsSubmitted && advance {
						poll++
						if poll >= 3 {
							accounts[second] = "Active"
						}
					}
					rows := []map[string]string{}
					for id, status := range accounts {
						if scenario == "swap" && advance && poll == 1 && id == second {
							continue
						} // Eventual visibility after PUT.
						rows = append(rows, map[string]string{"account_id": id, "status": status})
					}
					assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"addon": map[string]string{"uuid": addon}, "allowed_accounts": rows, "status": "Operational"}))
				default:
					t.Errorf("unexpected request %s", r.Method)
					w.WriteHeader(400)
				}
			}))
			defer srv.Close()
			client, err := api.New(config.PostgresBaseURL(srv.URL))
			require.NoError(t, err)
			meta := &Config{API: client, PrivatelinkAllowedAccountsAddVerifyTimeout: 1, PrivatelinkAllowedAccountsRemoveVerifyTimeout: 1}
			if scenario == "TerraformCLI no-refresh retry" {
				dir := t.TempDir()
				run := localTerraformCLI(t, dir, meta)
				source := fmt.Sprintf(`terraform {
  required_providers { herokux = { source = "davidji99/herokux" } }
}
resource "herokux_privatelink" "test" {
  addon_id = %q
  allowed_accounts = [%q, %q, %q]
}`, addon, keep, first, second)
				require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte(source), 0600))
				run(0, "import", "herokux_privatelink.test", addon)
				run(1, "apply", "-auto-approve", "-no-color")
				require.Equal(t, "Active", accounts[old])
				checkObserved := func(want []string, pending []string) {
					t.Helper()
					var state struct {
						Resources []struct {
							Instances []struct {
								Attributes struct {
									Observed []string `json:"observed_allowed_accounts"`
									Pending  []string `json:"pending_account_additions"`
								}
							}
						}
					}
					raw, readErr := os.ReadFile(filepath.Join(dir, "terraform.tfstate"))
					require.NoError(t, readErr)
					require.NoError(t, json.Unmarshal(raw, &state))
					attrs := state.Resources[0].Instances[0].Attributes
					require.ElementsMatch(t, want, attrs.Observed)
					require.ElementsMatch(t, pending, attrs.Pending)
				}
				checkObserved([]string{old, keep}, []string{first, second})
				mu.Lock()
				calls = nil
				mu.Unlock()
				run(2, "plan", "-refresh=false", "-detailed-exitcode", "-out=retry")
				mu.Lock()
				noPlanCalls := len(calls) == 0
				advance = true
				mu.Unlock()
				require.True(t, noPlanCalls)
				run(0, "apply", "retry")
				require.Equal(t, []string{"GET", "GET", "GET", "PATCH", "GET"}, calls)
				require.NotContains(t, accounts, old)
				checkObserved([]string{first, second, keep}, nil)
				run(0, "plan", "-refresh=false", "-detailed-exitcode")
				return
			}
			r := resourceHerokuxPrivatelink()
			d := r.Data(nil)
			d.SetId(addon)
			ds := r.ReadContext(t.Context(), d, meta)
			require.False(t, ds.HasError(), "%v", ds)
			state := d.State()
			desired := []any{keep, first, second}
			if scenario == "pure revocation" {
				desired = []any{keep}
			}
			cfg := terraform.NewResourceConfigRaw(map[string]any{"addon_id": addon, "allowed_accounts": desired})
			apply := func(wantError bool) {
				t.Helper()
				diff, planErr := r.Diff(t.Context(), state, cfg, meta)
				require.NoError(t, planErr)
				require.False(t, diff.Empty())
				if scenario == "activation lookup fails then retries" && wantError {
					// The protocol preserves a planned, known set even on errors. The retry marker must be unknown here.
					change := diff.Attributes["pending_account_additions.#"]
					require.NotNil(t, change)
					require.True(t, change.NewComputed)
				}
				next, applyDiags := r.Apply(t.Context(), state, diff, meta)
				require.Equal(t, wantError, applyDiags.HasError(), "%v", applyDiags)
				require.NotNil(t, next)
				state = next
			}
			calls = nil
			switch scenario {
			case "addition rejected":
				apply(true)
				require.Equal(t, []string{"GET", "PUT"}, calls)
				require.Equal(t, "Active", accounts[old])
				require.Equal(t, addon, state.ID)
				return
			case "activation lookup fails then retries":
				apply(true)
				require.Equal(t, []string{"GET", "PUT", "GET"}, calls)
				require.Equal(t, "Active", accounts[old])
				state, ds = r.RefreshWithoutUpgrade(t.Context(), state, meta)
				require.False(t, ds.HasError(), "%v", ds)
				// Both new accounts are now listed, but the second one still needs activation.
				calls = nil
				advance = true
				apply(false)
				require.Equal(t, []string{"GET", "GET", "GET", "PATCH", "GET"}, calls)
			case "swap":
				advance = true
				apply(false)
				require.Equal(t, []string{"GET", "PUT", "GET", "GET", "GET", "PATCH", "GET"}, calls)
			case "pure revocation":
				apply(false)
				require.Equal(t, []string{"GET", "PATCH", "GET"}, calls)
			}
			require.NotContains(t, accounts, old)
			require.Equal(t, "Provisioning", accounts[keep])
			diff, err := r.Diff(t.Context(), state, cfg, meta)
			require.NoError(t, err)
			require.True(t, diff.Empty())
		})
	}
}

func TestPrivatelinkSavedPlanDriftTerraformCLI(t *testing.T) {
	const addon = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	const keep, add, external = "100000000001", "100000000002", "100000000003"
	accounts := map[string]bool{keep: true}
	var writes []string
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			rows := []map[string]string{}
			for id := range accounts {
				rows = append(rows, map[string]string{"account_id": id, "status": "Active"})
			}
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"addon": map[string]string{"uuid": addon}, "allowed_accounts": rows, "status": "Operational"}))
			return
		}
		var body struct {
			Accounts []string `json:"allowed_accounts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		for _, id := range body.Accounts {
			writes = append(writes, r.Method+" "+id)
			switch r.Method {
			case http.MethodPut:
				accounts[id] = true
			case http.MethodPatch:
				delete(accounts, id)
			default:
				t.Errorf("unexpected method %s", r.Method)
			}
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, "{}")
	}))
	defer srv.Close()
	client, err := api.New(config.PostgresBaseURL(srv.URL))
	require.NoError(t, err)
	dir := t.TempDir()
	run := localTerraformCLI(t, dir, &Config{API: client, PrivatelinkAllowedAccountsAddVerifyTimeout: 1, PrivatelinkAllowedAccountsRemoveVerifyTimeout: 1})
	source := fmt.Sprintf(`terraform {
  required_providers { herokux = { source = "davidji99/herokux" } }
}
resource "herokux_privatelink" "test" {
  addon_id = %q
  allowed_accounts = [%q, %q]
}`, addon, keep, add)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte(source), 0600))
	run(0, "import", "herokux_privatelink.test", addon)
	run(0, "plan", "-out=add-only")
	mu.Lock()
	accounts[external] = true
	mu.Unlock()
	require.Contains(t, string(run(1, "apply", "-no-color", "add-only")), "fresh plan")
	mu.Lock()
	assert.Empty(t, writes)
	assert.True(t, accounts[external])
	mu.Unlock()

	plan := run(2, "plan", "-no-color", "-detailed-exitcode", "-out=reviewed")
	require.Contains(t, string(plan), external)
	run(0, "apply", "reviewed")
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []string{"PUT " + add, "PATCH " + external}, writes)
	require.Equal(t, map[string]bool{keep: true, add: true}, accounts)
}

func TestPrivatelinkVerificationDeadline(t *testing.T) {
	const addon = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	const response = `{"addon":{"uuid":"` + addon + `"},"allowed_accounts":[{"account_id":"100000000002","status":"Active"}]}`
	t.Run("deadline interrupts in-flight GET", func(t *testing.T) {
		aborted := make(chan struct{})
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
				close(aborted)
			case <-time.After(3 * time.Second):
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, response)
			}
		}))
		defer srv.Close()
		client, err := api.New(config.PostgresBaseURL(srv.URL))
		require.NoError(t, err)
		d := resourceHerokuxPrivatelink().Data(nil)
		d.SetId(addon)
		err = waitForPrivatelinkAccounts(t.Context(), d, client, 250*time.Millisecond, func(*postgres.Privatelink, *schema.Set) bool {
			t.Error("expired activation poll reached readiness check")
			return true
		})
		require.ErrorIs(t, err, context.DeadlineExceeded)
		select {
		case <-aborted:
		case <-time.After(2 * time.Second):
			t.Fatal("verification deadline did not cancel the server request")
		}
		require.NoError(t, t.Context().Err())
	})
	t.Run("cancel before accepting readiness", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, response)
		}))
		defer srv.Close()
		client, err := api.New(config.PostgresBaseURL(srv.URL))
		require.NoError(t, err)
		d := resourceHerokuxPrivatelink().Data(nil)
		d.SetId(addon)
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		err = waitForPrivatelinkAccounts(ctx, d, client, time.Minute, func(*postgres.Privatelink, *schema.Set) bool {
			cancel()
			return true
		})
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestPrivateLinkCreationWaitsForRequestedAccounts(t *testing.T) {
	const addon = "6fae1ee0-c034-4775-a798-890bc64f98eb"
	body := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	defer srv.Close()
	client, err := api.New(config.PostgresBaseURL(srv.URL))
	require.NoError(t, err)
	desired := schema.NewSet(schema.HashString, []any{"123456789012"})
	poll := privatelinkProvisioning(t.Context(), client, addon, desired)
	for _, tc := range []struct{ status, accounts, want string }{
		{"Provisioning", "null", "Provisioning"},
		{"Operational", "[]", "Provisioning"},
		{"Operational", `[{"account_id":"999999999999","status":"Active"}]`, "Provisioning"},
		{"Operational", `[{"account_id":"123456789012","status":"Unknown"}]`, "Provisioning"},
		{"Operational", `[{"account_id":"123456789012","status":"Active"}]`, "Operational"},
		{"Deprovisioned", "[]", "Deprovisioned"},
	} {
		body = fmt.Sprintf(`{"addon":{"uuid":%q},"status":%q,"allowed_accounts":%s}`, addon, tc.status, tc.accounts)
		_, state, err := poll()
		require.NoError(t, err)
		require.Equal(t, tc.want, state)
	}
}
