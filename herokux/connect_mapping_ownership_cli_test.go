package herokux

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"testing"

	"github.com/stretchr/testify/require"
)

// Only Terraform Core applies ignore_changes and upgrades saved state before a no-refresh destroy.
func TestConnectMappingOwnershipTerraformCLI(t *testing.T) {
	tf := os.Getenv("TF_ACC_TERRAFORM_PATH")
	if tf == "" {
		t.Skip("set TF_ACC_TERRAFORM_PATH to a Terraform executable")
	}
	const addr = "herokux_connect_mappings.test"
	const account = `jsonencode({mappings = [{object_name = "Account"}]})`
	for _, scenario := range []string{"import", "upgrade and adopt", "upgrade without ownership marker", "upgrade and destroy without refresh"} {
		t.Run(scenario, func(t *testing.T) {
			f := newConnectMappingFixture(t, map[string]string{"Account": "account-id", "Contact": "contact-id"})
			dir := t.TempDir()
			run := localTerraformCLI(t, dir, f.meta)
			writeConfig := func(mappings, lifecycle string) {
				t.Helper()
				hcl := fmt.Sprintf(`terraform {
  required_providers {
    herokux = { source = "davidji99/herokux" }
  }
}
resource "herokux_connect_mappings" "test" {
  app_id = %q
  connect_id = %q
  mappings = %s
  %s
}`, connectFixtureApp, connectFixtureID, mappings, lifecycle)
				require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte(hcl), 0600))
			}

			stateFile := filepath.Join(dir, "terraform.tfstate")
			assertState := func(owned bool) {
				t.Helper()
				raw, err := os.ReadFile(stateFile)
				require.NoError(t, err)
				var state struct {
					Resources []struct {
						Type      string
						Instances []struct{ Attributes map[string]any }
					}
				}
				require.NoError(t, json.Unmarshal(raw, &state))
				var attrs map[string]any
				for _, resource := range state.Resources {
					if resource.Type == "herokux_connect_mappings" {
						require.Len(t, resource.Instances, 1)
						attrs = resource.Instances[0].Attributes
					}
				}
				require.NotNil(t, attrs)
				require.Equal(t, owned, attrs["ownership_recorded"])
				if owned {
					require.Equal(t, map[string]any{"Account": "account-id"}, attrs["mapping_data"])
				} else {
					require.JSONEq(t, emptyConnectMappings, attrs["mappings"].(string))
					require.Empty(t, attrs["mapping_data"])
				}
			}
			writeConfig(account, "lifecycle { ignore_changes = [mappings] }")
			if scenario == "import" {
				run(0, "import", "-input=false", addr, connectFixtureApp+":"+connectFixtureID)
				run(0, "apply", "-input=false", "-auto-approve")
				assertState(false)
			} else {
				marker := `"ownership_recorded": true,`
				if scenario == "upgrade without ownership marker" {
					marker = ""
				}
				state := fmt.Sprintf(connectMappingsV0StateFormat, marker)
				require.NoError(t, os.WriteFile(stateFile, []byte(state), 0600))
			}
			if scenario == "upgrade and destroy without refresh" {
				run(0, "destroy", "-refresh=false", "-auto-approve")
				require.Empty(t, f.writes)
				require.Equal(t, map[string]string{"Account": "account-id", "Contact": "contact-id"}, f.ids)
				return
			}
			run(0, "plan", "-input=false", "-detailed-exitcode")
			require.Empty(t, f.writes)
			writeConfig(`jsonencode({version = 1, mappings = []})`, "")
			run(0, "plan", "-input=false", "-detailed-exitcode")
			if scenario == "upgrade and adopt" {
				// The computed output becomes known only after the built-in resource is applied.
				source := `resource "terraform_data" "document" { input = ` + account + ` }`
				require.NoError(t, os.WriteFile(filepath.Join(dir, "document.tf"), []byte(source), 0600))
				writeConfig("terraform_data.document.output", "")
			} else {
				writeConfig(account, "")
			}
			const consumer = `resource "terraform_data" "consumer" {
  for_each = toset(["Account"])
  input    = herokux_connect_mappings.test.mapping_data[each.key]
}
output "direct_id"   { value = herokux_connect_mappings.test.mapping_data["Account"] }
output "consumed_id" { value = terraform_data.consumer["Account"].input }
`
			require.NoError(t, os.WriteFile(filepath.Join(dir, "consumer.tf"), []byte(consumer), 0600))
			run(0, "plan", "-input=false", "-out=adopt.tfplan")
			var plan struct {
				OutputChanges map[string]struct {
					AfterUnknown bool `json:"after_unknown"`
				} `json:"output_changes"`
			}
			require.NoError(t, json.Unmarshal(run(0, "show", "-json", "adopt.tfplan"), &plan))
			for _, name := range []string{"direct_id", "consumed_id"} {
				require.True(t, plan.OutputChanges[name].AfterUnknown, "%s must wait for the verified mapping ID", name)
			}
			require.Empty(t, f.writes)
			run(0, "apply", "-input=false", "adopt.tfplan")
			assertState(true)
			require.Equal(t, "account-id", string(run(0, "output", "-raw", "direct_id")))
			require.Equal(t, "account-id", string(run(0, "output", "-raw", "consumed_id")))
			require.Equal(t, []string{"import Account"}, f.writes)
			run(0, "plan", "-input=false", "-detailed-exitcode")
			run(0, "destroy", "-auto-approve")
			require.Equal(t, []string{"import Account", "delete account-id"}, f.writes)
			require.Equal(t, map[string]string{"Contact": "contact-id"}, f.ids)
		})
	}
}

const connectMappingsV0StateFormat = `{
  "version": 4, "terraform_version": "1.13.5", "serial": 1, "lineage": "f630a412-45a6-4cd8-bb99-0e5b7b763731",
  "outputs": {},
  "resources": [{
    "mode": "managed", "type": "herokux_connect_mappings", "name": "test",
    "provider": "provider[\"registry.terraform.io/davidji99/herokux\"]",
    "instances": [{"schema_version": 0, "attributes": {
      "id": "` + connectFixtureID + `", "connect_id": "` + connectFixtureID + `", "app_id": "` + connectFixtureApp + `",
      %s
      "mappings": "{\"mappings\":[{\"object_name\":\"Account\"},{\"object_name\":\"Contact\"}]}",
      "mapping_data": {"Account":"account-id","Contact":"contact-id"},
      "mapping_object_names": ["Account","Contact"], "mapping_ids": ["account-id","contact-id"]
    }}]
  }]
}`

func TestConnectInterruptedImportTerraformCLI(t *testing.T) {
	f := newConnectMappingFixture(t, map[string]string{"Account": "account-id", "Contact": "external"})
	dir := t.TempDir()
	run := localTerraformCLI(t, dir, f.meta)
	const addr = "herokux_connect_mappings.test"
	source := fmt.Sprintf(`terraform {
  required_providers { herokux = { source = "davidji99/herokux" } }
}
resource "herokux_connect_mappings" "test" {
  app_id = %q
  connect_id = %q
  mappings = jsonencode({mappings = [{object_name="Account"},{object_name="Lead"}]})
}`, connectFixtureApp, connectFixtureID)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte(source), 0600))
	f.afterImportInventoryBody = `{}`
	require.Contains(t, string(run(1, "apply", "-auto-approve", "-no-color")), "interrupted import")
	require.Equal(t, []string{"import Account", "import Lead"}, f.writes)
	var state struct {
		Resources []struct {
			Instances []struct{ Attributes map[string]any }
		}
	}
	raw, err := os.ReadFile(filepath.Join(dir, "terraform.tfstate"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &state))
	attrs := state.Resources[0].Instances[0].Attributes
	require.Equal(t, connectFixtureID, attrs["id"])
	require.Equal(t, map[string]any{"Account": "account-id", "Lead": ""}, attrs["pending_mapping_ids"])
	f.mu.Lock()
	f.inventoryBody, f.afterImportInventoryBody = "", ""
	f.ids["Lead"] = "other-id"
	f.mu.Unlock()
	require.Contains(t, string(run(1, "destroy", "-auto-approve", "-no-color")), "interrupted import")
	require.Equal(t, addr, strings.TrimSpace(string(run(0, "state", "list"))))
	require.Equal(t, []string{"import Account", "import Lead"}, f.writes)
	run(0, "state", "rm", addr)
	run(0, "import", addr, connectFixtureApp+":"+connectFixtureID)
	run(0, "apply", "-auto-approve")
	run(0, "destroy", "-auto-approve")
	require.Equal(t, map[string]string{"Contact": "external"}, f.ids)
}
