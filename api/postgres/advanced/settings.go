package advanced

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
)

func (c *Client) GetSettings(ctx context.Context, db string) (map[string]string, error) {
	var response struct {
		Items []struct {
			Name    string          `json:"name"`
			Current json.RawMessage `json:"current"`
		} `json:"items"`
	}
	if err := c.request(ctx, http.MethodGet, databasePath(db)+"/settings", nil, &response); err != nil {
		return nil, err
	}
	if response.Items == nil {
		return nil, fmt.Errorf("missing items in Postgres Advanced settings response")
	}
	values := make(map[string]string)
	for _, item := range response.Items {
		if item.Name == "" || len(item.Current) == 0 {
			return nil, fmt.Errorf("missing name or current value in Postgres Advanced settings response")
		}
		value := string(item.Current)
		if value == "null" {
			continue
		}
		switch item.Current[0] {
		case '"':
			if err := json.Unmarshal(item.Current, &value); err != nil {
				return nil, err
			}
		case '{', '[':
			return nil, fmt.Errorf("unexpected non-scalar current value for setting %q", item.Name)
		}
		values[item.Name] = value
	}
	return values, nil
}

type settingRule struct {
	choices   []string
	foldCase  bool
	minMillis int64 // Used for duration settings, which have no choices.
}

var booleanSetting = settingRule{
	choices: []string{"true", "false", "on", "off", "yes", "no", "1", "0"}, foldCase: true,
}

var settingRules = map[string]settingRule{
	"auto_explain.log_analyze":            booleanSetting,
	"auto_explain.log_buffers":            booleanSetting,
	"auto_explain.log_format":             {choices: []string{"text", "xml", "json", "yaml"}},
	"auto_explain.log_min_duration":       {minMillis: -1},
	"auto_explain.log_nested_statements":  booleanSetting,
	"auto_explain.log_triggers":           booleanSetting,
	"auto_explain.log_verbose":            booleanSetting,
	"idle_in_transaction_session_timeout": {minMillis: 0},
	"log_connections":                     booleanSetting,
	"log_lock_waits":                      booleanSetting,
	"log_min_duration_statement":          {minMillis: -1},
	"log_min_error_statement":             {choices: []string{"error", "log", "fatal", "panic"}},
	"log_statement":                       {choices: []string{"none", "ddl", "mod", "all"}},
	"track_functions":                     {choices: []string{"none", "pl", "all"}},
}

// ValidateSettingName also works when Terraform hasn't resolved the value yet.
func ValidateSettingName(name string) error {
	if _, ok := settingRules[name]; !ok {
		return fmt.Errorf("unknown Postgres setting %q (supported settings are %s)", name, strings.Join(slices.Sorted(maps.Keys(settingRules)), ", "))
	}
	return nil
}

// ValidateSetting checks a configured value before it is sent to the Advanced API.
func ValidateSetting(name, value string) error {
	if err := ValidateSettingName(name); err != nil {
		return err
	}
	rule := settingRules[name]
	if rule.choices != nil {
		candidate := value
		if rule.foldCase {
			candidate = strings.ToLower(value)
		}
		if !slices.Contains(rule.choices, candidate) {
			return fmt.Errorf("invalid value %q for Postgres setting %q (expected one of %s)", value, name, strings.Join(rule.choices, ", "))
		}
		return nil
	}
	ms, err := strconv.ParseInt(value, 10, 32)
	if err != nil || ms < rule.minMillis || strconv.FormatInt(ms, 10) != value {
		return fmt.Errorf("invalid value %q for Postgres setting %q (expected whole milliseconds from %d to %d in canonical decimal form)", value, name, rule.minMillis, math.MaxInt32)
	}
	return nil
}

func (c *Client) UpdateSettings(ctx context.Context, db string, changes map[string]string) error {
	pairs := make([]string, 0, len(changes))
	for name, value := range changes {
		if err := ValidateSetting(name, value); err != nil {
			return err
		}
		pairs = append(pairs, name+":"+value)
	}
	sort.Strings(pairs)
	return c.request(ctx, http.MethodPut, databasePath(db)+"/settings", map[string]string{"settings": strings.Join(pairs, ",")}, nil)
}
