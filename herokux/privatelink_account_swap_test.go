package herokux

import (
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
