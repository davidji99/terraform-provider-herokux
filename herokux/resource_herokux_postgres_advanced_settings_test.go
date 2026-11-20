package herokux

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestPostgresAdvancedSettings(t *testing.T) {
	f := newAdvancedTestResource(t, "herokux_postgres_advanced_settings")
	require.NotNil(t, f.resource, "provider must register Advanced settings")
	const path = "/settings"
	const initial = `{"items":[{"name":"log_lock_waits","current":false},{"name":"log_min_duration_statement","current":0},{"name":"auto_explain.log_format","current":"text"},{"name":"log_statement","current":"ddl"},{"name":"log_min_error_statement","current":"info"},{"name":"auto_explain.log_verbose","current":null}]}`
	const batch = `{"settings":"auto_explain.log_format:text,log_lock_waits:off,log_min_duration_statement:0"}`
	settings := map[string]any{"log_lock_waits": "off", "log_min_duration_statement": "0", "auto_explain.log_format": "text"}
	args := map[string]any{"postgres_id": advancedTestDB, "settings": settings}
	f.expect("GET", path, "", 200, `{"items":[{"name":"log_lock_waits","current":true},{"name":"log_min_duration_statement","current":500},{"name":"auto_explain.log_format","current":null}]}`)
	f.expect("PUT", path, batch, 200, `{"changes":[]}`)
	f.expect("GET", path, "", 200, initial)
	f.apply(args)
	require.True(t, f.plan(args).Empty())
	require.NotContains(t, f.state.Attributes, "settings.log_statement")

	f.expect("GET", path, "", 200, initial)
	f.importID(advancedTestDB)
	require.Equal(t, "false", f.state.Attributes["settings.log_lock_waits"])
	require.Equal(t, "ddl", f.state.Attributes["settings.log_statement"])
	require.Equal(t, "info", f.state.Attributes["settings.log_min_error_statement"])
	require.NotContains(t, f.state.Attributes, "settings.auto_explain.log_verbose")

	// Drop the imported log_statement from management without resetting it.
	f.expect("GET", path, "", 200, initial)
	f.apply(args)
	require.True(t, f.plan(args).Empty())

	settings["auto_explain.log_format"] = "json"
	const change = `{"settings":"auto_explain.log_format:json"}`
	updated := strings.Replace(initial, `"text"`, `"json"`, 1)
	f.expect("GET", path, "", 200, initial)
	f.expect("PUT", path, change, 200, `{"changes":[{"name":"auto_explain.log_format","previous":"text","current":"json"}]}`)
	f.expect("GET", path, "", 200, initial) // Heroku accepted the change before it became visible.
	f.expect("GET", path, "", 200, updated)
	f.apply(args)
	require.True(t, f.plan(args).Empty())

	f.expect("GET", path, "", 200, initial)
	f.refresh()
	require.False(t, f.plan(args).Empty())
	f.expect("GET", path, "", 200, initial)
	f.expect("PUT", path, change, 422, `{"message":"database busy"}`)
	prior, ds := f.resource.Apply(t.Context(), f.state, f.plan(args), f.meta)
	require.True(t, ds.HasError())
	require.Equal(t, "text", prior.Attributes["settings.auto_explain.log_format"])
	f.expect("GET", path, "", 200, initial)
	f.expect("PUT", path, change, 200, `{"changes":[]}`)
	f.expect("GET", path, "", 200, updated)
	f.apply(args)
	require.True(t, f.plan(args).Empty())
	f.destroy()
}

func TestPostgresAdvancedSettingsValidation(t *testing.T) {
	r := resourceHerokuxPostgresAdvancedSettings()
	require.Len(t, r.ValidateRawResourceConfigFuncs, 1)
	validate := func(settings cty.Value) *schema.ValidateResourceConfigFuncResponse {
		t.Helper()
		req := schema.ValidateResourceConfigFuncRequest{
			RawConfig: cty.ObjectVal(map[string]cty.Value{"settings": settings}),
		}
		resp := &schema.ValidateResourceConfigFuncResponse{}
		r.ValidateRawResourceConfigFuncs[0](t.Context(), req, resp)
		return resp
	}
	unknown := cty.UnknownVal(cty.String)
	require.Empty(t, validate(cty.UnknownVal(cty.Map(cty.String))).Diagnostics)
	require.Empty(t, validate(cty.MapVal(map[string]cty.Value{
		"log_statement": unknown, "log_connections": cty.NullVal(cty.String),
	})).Diagnostics)

	d := validate(cty.MapVal(map[string]cty.Value{
		"log_statment": unknown, "track_functions": cty.StringVal("sometimes"), "log_statement": unknown,
	})).Diagnostics
	require.Len(t, d, 2)
	require.Equal(t, cty.GetAttrPath("settings").IndexString("log_statment"), d[0].AttributePath)
	require.Contains(t, d[0].Summary, "supported settings")
	require.Equal(t, cty.GetAttrPath("settings").IndexString("track_functions"), d[1].AttributePath)
	require.Contains(t, d[1].Summary, "none, pl, all")
}

func TestPostgresAdvancedSettingsRejectInvalidWrite(t *testing.T) {
	f := newAdvancedTestResource(t, "herokux_postgres_advanced_settings")
	err := f.meta.API.PostgresAdvanced.UpdateSettings(t.Context(), advancedTestDB, map[string]string{
		"log_statement": "ddl", "log_min_error_statement": "info",
	})
	require.ErrorContains(t, err, "error, log, fatal, panic")
}

func TestPostgresAdvancedSettingsRefreshErrors(t *testing.T) {
	f := newAdvancedTestResource(t, "herokux_postgres_advanced_settings")
	f.expect("GET", "/settings", "", 200, `{"items":[{"name":"track_functions","current":"pl"}]}`)
	f.importID(advancedTestDB)
	for _, bad := range []struct {
		status int
		body   string
	}{
		{403, `{"message":"forbidden"}`}, {500, `{}`}, {200, `{"changes":[]}`},
		{200, `{"items":[{"name":"track_functions"}]}`},
	} {
		f.expect("GET", "/settings", "", bad.status, bad.body)
		_, ds := f.resource.RefreshWithoutUpgrade(t.Context(), f.state, f.meta)
		require.True(t, ds.HasError())
	}
	f.expect("GET", "/settings", "", 404, `{}`)
	f.refresh()
	require.Nil(t, f.state)
}
