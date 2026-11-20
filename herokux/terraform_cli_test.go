package herokux

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/stretchr/testify/require"
)

// Start Terraform with a provider connected to the test's local API fixture.
func localTerraformCLI(t *testing.T, dir string, meta *Config) func(int, ...string) []byte {
	t.Helper()
	tf := os.Getenv("TF_ACC_TERRAFORM_PATH")
	if tf == "" {
		t.Skip("set TF_ACC_TERRAFORM_PATH to a Terraform executable")
	}
	ctx, stop := context.WithCancel(t.Context())
	t.Cleanup(stop)
	const address = "registry.terraform.io/davidji99/herokux"
	info, closed, err := plugin.DebugServe(ctx, &plugin.ServeOpts{
		ProviderAddr: address, UseTFLogSink: t, NoLogOutputOverride: true,
		ProviderFunc: func() *schema.Provider {
			p := New()
			p.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) { return meta, nil }
			return p
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { stop(); <-closed })
	attach, err := json.Marshal(map[string]plugin.ReattachConfig{address: info})
	require.NoError(t, err)
	rc := filepath.Join(dir, "terraformrc")
	require.NoError(t, os.WriteFile(rc, nil, 0600))
	return func(wantExit int, args ...string) []byte {
		t.Helper()
		cmd := exec.CommandContext(ctx, tf, args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "TF_REATTACH_PROVIDERS="+string(attach), "TF_CLI_CONFIG_FILE="+rc, "CHECKPOINT_DISABLE=1", "TF_IN_AUTOMATION=1")
		out, err := cmd.CombinedOutput()
		require.NotNil(t, cmd.ProcessState, "could not run Terraform: %v", err)
		require.Equal(t, wantExit, cmd.ProcessState.ExitCode(), "%s: %s", args, out)
		return out
	}
}
