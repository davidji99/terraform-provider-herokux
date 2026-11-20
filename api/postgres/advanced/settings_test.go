package advanced

import (
	"maps"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSetting(t *testing.T) {
	tests := []struct {
		names    []string
		valid    []string
		invalid  []string
		expected string
	}{
		{
			[]string{"log_connections", "log_lock_waits", "auto_explain.log_analyze", "auto_explain.log_buffers",
				"auto_explain.log_nested_statements", "auto_explain.log_triggers", "auto_explain.log_verbose"},
			[]string{"true", "false", "on", "off", "yes", "no", "1", "0", "TRUE", "Off", "YES", "nO"},
			[]string{"t", "enable", "2", " on"}, "true, false, on, off, yes, no, 1, 0",
		},
		{
			[]string{"log_min_duration_statement", "auto_explain.log_min_duration"},
			[]string{"-1", "0", "500", "2147483647"},
			[]string{"-2", "2147483648", "5ms", "5.0", "+5", "05", "-0"}, "-1 to 2147483647",
		},
		{
			[]string{"idle_in_transaction_session_timeout"},
			[]string{"0", "500", "2147483647"}, []string{"-1", "2147483648", "0s", "00"}, "0 to 2147483647",
		},
		{[]string{"auto_explain.log_format"}, []string{"text", "xml", "json", "yaml"}, []string{"csv"}, "text, xml, json, yaml"},
		{[]string{"log_min_error_statement"}, []string{"error", "log", "fatal", "panic"}, []string{"info"}, "error, log, fatal, panic"},
		{[]string{"log_statement"}, []string{"none", "ddl", "mod", "all"}, []string{"select", "DDL", " ddl "}, "none, ddl, mod, all"},
		{[]string{"track_functions"}, []string{"none", "pl", "all"}, []string{"sometimes"}, "none, pl, all"},
	}

	var tested []string
	for _, tc := range tests {
		for _, name := range tc.names {
			tested = append(tested, name)
			t.Run(name, func(t *testing.T) {
				for _, value := range tc.valid {
					require.NoError(t, ValidateSetting(name, value), "value %q", value)
				}
				for _, value := range append(slices.Clone(tc.invalid), "", "ddl,track_functions:all") {
					err := ValidateSetting(name, value)
					require.ErrorContains(t, err, name)
					require.ErrorContains(t, err, tc.expected)
				}
			})
		}
	}
	require.ElementsMatch(t, tested, slices.Collect(maps.Keys(settingRules)))
	slices.Sort(tested)
	for _, name := range []string{"log_statment", "auto_explain.log-format", "log_statement,track_functions"} {
		err := ValidateSetting(name, "all")
		require.ErrorContains(t, err, name)
		require.ErrorContains(t, err, strings.Join(tested, ", "))
	}
}

func TestSettingReferenceMatchesCatalog(t *testing.T) {
	doc, err := os.ReadFile("../../../docs/resources/postgres_advanced_settings.md")
	require.NoError(t, err)
	rows := regexp.MustCompile("(?m)^\\| `([^`]+)` \\|").FindAllStringSubmatch(string(doc), -1)
	var documented []string
	for _, row := range rows {
		documented = append(documented, row[1])
	}
	require.ElementsMatch(t, slices.Collect(maps.Keys(settingRules)), documented)
}
