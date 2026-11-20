package herokux

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/applink"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

// Run against Terraform itself: the SDK's ResourceData unit-test helpers do not model ephemeral values.
func TestAppLinkTerraformCLI(t *testing.T) {
	tf := os.Getenv("TF_ACC_TERRAFORM_PATH")
	if tf == "" {
		t.Skip("set TF_ACC_TERRAFORM_PATH to a Terraform 1.11+ executable")
	}
	f := newAppLinkFixture(t)
	dir := t.TempDir()
	t.Setenv("TF_ACC_LOG_PATH", filepath.Join(dir, "sdk.log"))
	providerLog, err := os.Create(filepath.Join(dir, "provider.log"))
	require.NoError(t, err)
	oldLog := log.Writer()
	log.SetOutput(providerLog)
	t.Cleanup(func() { log.SetOutput(oldLog); _ = providerLog.Close() })

	ctx, stop := context.WithCancel(t.Context())
	defer stop()
	const address = "registry.terraform.io/davidji99/herokux"
	const instance = "herokux_applink_connection.test"
	info, closed, err := plugin.DebugServe(ctx, &plugin.ServeOpts{ProviderAddr: address, UseTFLogSink: t, NoLogOutputOverride: true, ProviderFunc: func() *schema.Provider {
		p := New()
		p.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) { return f.meta, nil }
		return p
	}})
	require.NoError(t, err)
	t.Cleanup(func() { stop(); <-closed })
	reattach, err := json.Marshal(map[string]plugin.ReattachConfig{address: info})
	require.NoError(t, err)

	args := map[string]any{"app_id": linkApp, "addon_id": linkAddon, "name": linkName}
	source := map[string]any{
		"terraform": map[string]any{"required_providers": map[string]any{"herokux": map[string]string{"source": "davidji99/herokux"}}},
		"variable":  map[string]any{"ci_key": map[string]any{"type": "string", "sensitive": true, "ephemeral": true, "default": nil}},
		"resource":  map[string]any{"herokux_applink_connection": map[string]any{"test": args}},
	}
	setCredentials := func() {
		args["client_id"] = "consumer"
		args["username"] = "integration@example.org"
		args["jwt_private_key_wo"] = "${var.ci_key}"
	}
	key := f.key
	marker := []byte(strings.Split(key, "\n")[1])
	secretFree := func(label string, b []byte) {
		t.Helper()
		require.False(t, bytes.Contains(b, marker), "private key material in %s", label)
	}
	run := func(exit int, argv ...string) string {
		t.Helper()
		data, encErr := json.Marshal(source)
		require.NoError(t, encErr)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.tf.json"), data, 0600))
		cmd := exec.CommandContext(ctx, tf, argv...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "TF_REATTACH_PROVIDERS="+string(reattach), "TF_VAR_ci_key="+key, "TF_CLI_CONFIG_FILE="+filepath.Join(dir, "terraformrc"), "CHECKPOINT_DISABLE=1", "TF_LOG=TRACE", "TF_LOG_PATH="+filepath.Join(dir, "terraform.log"), "TF_IN_AUTOMATION=1")
		out, runErr := cmd.CombinedOutput()
		secretFree("CLI output and diagnostics", out)
		code := 0
		if runErr != nil {
			require.NotNil(t, cmd.ProcessState, "could not start Terraform: %v", runErr)
			code = cmd.ProcessState.ExitCode()
		}
		require.Equal(t, exit, code, "%s: %s", argv, out)
		for _, file := range []string{"terraform.tfstate", "terraform.tfstate.backup"} {
			b, stateErr := os.ReadFile(filepath.Join(dir, file))
			if !os.IsNotExist(stateErr) {
				require.NoError(t, stateErr)
				secretFree(file, b)
			}
		}
		return string(out)
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "terraformrc"), nil, 0600))
	require.Contains(t, run(1, "plan", "-input=false", "-no-color"), "required to create")
	setCredentials()
	run(0, "plan", "-input=false", "-out=saved.tfplan", "-no-color")
	run(0, "show", "-json", "saved.tfplan")
	archive, err := zip.OpenReader(filepath.Join(dir, "saved.tfplan"))
	require.NoError(t, err)
	for _, entry := range archive.File {
		rd, openErr := entry.Open()
		require.NoError(t, openErr)
		content, readErr := io.ReadAll(rd)
		_ = rd.Close()
		require.NoError(t, readErr)
		secretFree("plan entry "+entry.Name, content)
	}
	require.NoError(t, archive.Close())
	run(0, "apply", "-input=false", "-no-color", "saved.tfplan")
	run(0, "plan", "-input=false", "-detailed-exitcode", "-no-color")
	require.Equal(t, 1, f.creates)

	validKey := key
	key = strings.Replace(key, "PRIVATE KEY", "INVALID KEY", 1)
	require.Contains(t, run(1, "plan", "-replace="+instance, "-input=false", "-no-color"), "unencrypted RSA private key")
	require.Zero(t, f.deletes)
	key = validKey
	key += "\n" // a different write-only value cannot cause a reconnect
	run(0, "plan", "-input=false", "-detailed-exitcode", "-no-color")
	args["client_id"] = "next-consumer"
	args["app_id"] = linkOtherApp
	run(0, "apply", "-auto-approve", "-input=false", "-no-color")
	require.Equal(t, 1, f.creates)
	require.Zero(t, f.deletes)
	require.Contains(t, run(0, "state", "show", "-no-color", instance), linkOtherApp+":"+linkAddon+":"+linkName)

	run(0, "state", "rm", instance)
	rotationKey := key
	key = "" // the import and subsequent refresh have no PEM
	for _, k := range []string{"client_id", "username", "jwt_private_key_wo"} {
		delete(args, k)
	}
	args["app_id"] = linkApp
	run(0, "import", "-input=false", "-no-color", instance, linkID)
	run(0, "plan", "-detailed-exitcode", "-input=false", "-no-color")
	require.Equal(t, 1, f.creates)
	require.Zero(t, f.deletes)
	require.Contains(t, run(1, "plan", "-replace="+instance, "-input=false", "-no-color"), "required to create")
	setCredentials()
	key = rotationKey
	f.Lock()
	f.key = key
	f.Unlock()
	run(0, "apply", "-auto-approve", "-input=false", "-replace="+instance, "-no-color")
	require.Equal(t, 2, f.creates)
	require.Equal(t, 1, f.deletes)
	require.Contains(t, run(0, "state", "show", "-no-color", instance), "remote-2")

	f.Lock()
	f.status = "disconnected"
	f.Unlock()
	delete(args, "jwt_private_key_wo")
	require.Contains(t, run(1, "plan", "-input=false", "-no-color"), "required to create")
	setCredentials()
	run(0, "apply", "-auto-approve", "-input=false", "-no-color")
	require.Equal(t, 3, f.creates)
	require.Equal(t, 1, f.deletes)
	require.Contains(t, run(0, "state", "show", "-no-color", instance), "remote-3")

	f.Lock()
	f.serviceStatus = 403
	f.Unlock()
	require.Contains(t, run(1, "plan", "-input=false", "-no-color"), "HTTP 403")
	f.Lock()
	f.serviceStatus = 0
	f.Unlock()
	run(0, "destroy", "-auto-approve", "-input=false", "-no-color")
	require.Equal(t, 2, f.deletes)
	f.Lock()
	f.afterCreate = "failed"
	f.Unlock()
	require.Contains(t, run(1, "apply", "-auto-approve", "-input=false", "-no-color"), "failed")
	require.Contains(t, run(0, "state", "show", "-no-color", instance), "remote-4")
	run(0, "destroy", "-auto-approve", "-input=false", "-no-color")
	require.Equal(t, 3, f.deletes)
	for _, name := range []string{"terraform.tfstate", "terraform.tfstate.backup", "provider.log", "sdk.log", "terraform.log"} {
		b, readErr := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, readErr)
		secretFree(name, b)
	}
}

func TestAccHerokuxAppLinkConnection(t *testing.T) {
	testAccConfig.SkipUnlessAccTest(t)
	for _, k := range []string{"HEROKUX_APPLINK_APP_ID", "HEROKUX_APPLINK_ADDON_ID", "HEROKUX_APPLINK_CLIENT_ID", "HEROKUX_APPLINK_USERNAME", "TF_VAR_applink_test_key"} {
		if os.Getenv(k) == "" {
			t.Skipf("set %s to use a dedicated AppLink/Salesforce sandbox fixture", k)
		}
	}
	testAccPreCheck(t)
	const addr = "herokux_applink_connection.test"
	name := "tftest" + acctest.RandString(10)
	client, err := api.New(config.APIToken(os.Getenv("HEROKU_API_KEY")))
	require.NoError(t, err)
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{Config: liveAppLinkConfig(name), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(addr, "status", "connected"), resource.TestCheckResourceAttrSet(addr, "connection_id"))},
			{ResourceName: addr, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"client_id", "username", "login_url"}},
		},
		CheckDestroy: func(_ *terraform.State) error {
			s, sessionErr := client.AppLink.Session(t.Context(), os.Getenv("HEROKUX_APPLINK_APP_ID"), os.Getenv("HEROKUX_APPLINK_ADDON_ID"))
			if sessionErr != nil {
				return sessionErr
			}
			c, lookupErr := s.Get(t.Context(), name)
			if errors.Is(lookupErr, applink.ErrNotFound) {
				return nil
			}
			if lookupErr != nil {
				return lookupErr
			}
			if c.Status == "disconnected" {
				return nil
			}
			return fmt.Errorf("test connection %s is still %s", name, c.Status)
		},
	})
}

func liveAppLinkConfig(name string) string {
	return fmt.Sprintf(`
variable "applink_test_key" {
  type = string
  sensitive = true
  ephemeral = true
}
resource "herokux_applink_connection" "test" {
  app_id             = %q
  addon_id           = %q
  name               = %q
  client_id          = %q
  username           = %q
  login_url          = "https://test.salesforce.com"
  jwt_private_key_wo = var.applink_test_key
}`, os.Getenv("HEROKUX_APPLINK_APP_ID"), os.Getenv("HEROKUX_APPLINK_ADDON_ID"), name, os.Getenv("HEROKUX_APPLINK_CLIENT_ID"), os.Getenv("HEROKUX_APPLINK_USERNAME"))
}
