package advanced

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
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

var settingName = regexp.MustCompile(`^[a-z_][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)

// ValidateSetting checks that one entry can be encoded as a CLI name:value pair.
func ValidateSetting(name, value string) error {
	if !settingName.MatchString(name) {
		return fmt.Errorf("invalid Postgres setting name %q", name)
	}
	if value == "" || strings.TrimSpace(value) != value || strings.ContainsAny(value, ",:\r\n") {
		return fmt.Errorf("invalid value for Postgres setting %q: use a nonempty value without surrounding whitespace, commas, colons or newlines", name)
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
