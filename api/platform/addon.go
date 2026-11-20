package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	heroku "github.com/davidji99/heroku-go/v5"
)

type Addon struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
	App  *App    `json:"app,omitempty"`
}

// PostgresCredentialApps lists the distinct apps using this credential across all attachment pages.
func (p *Platform) PostgresCredentialApps(ctx context.Context, addon, credential string) ([]string, error) {
	path := p.http.RequestURL("/addons/%s/addon-attachments", url.PathEscape(addon))
	var apps []string
	seen := map[string]bool{}
	for next := ""; ; {
		req := p.http.R().SetContext(ctx).SetHeader("Accept", DefaultAcceptHeader)
		if next != "" {
			req.SetHeader("Range", next)
		}
		resp, err := req.Get(path)
		if err != nil {
			return nil, fmt.Errorf("list attachments for %s: %w", addon, err)
		}
		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusPartialContent {
			return nil, fmt.Errorf("list attachments for %s: HTTP %d", addon, resp.StatusCode())
		}
		var page []struct {
			heroku.AddOnAttachment
			NamespaceConfig struct {
				Cred string `json:"cred"`
				Role string `json:"role"`
			} `json:"namespace_config"`
		}
		if err := json.Unmarshal(resp.Body(), &page); err != nil || page == nil {
			return nil, fmt.Errorf("invalid attachment list for %s", addon)
		}
		for _, a := range page {
			app := a.App.Name
			if app == "" {
				app = a.App.ID
			}
			if app == "" {
				return nil, fmt.Errorf("attachment on %s is missing its app identity", addon)
			}
			if a.NamespaceConfig.Cred == credential || a.NamespaceConfig.Role == credential ||
				(a.Namespace != nil && (*a.Namespace == "credential:"+credential || *a.Namespace == "role:"+credential)) {
				apps = append(apps, app)
			}
		}
		next = resp.Header().Get("Next-Range")
		if next == "" && resp.StatusCode() == http.StatusOK {
			break
		}
		if next == "" || seen[next] {
			return nil, fmt.Errorf("missing or repeated attachment pagination range for %s", addon)
		}
		seen[next] = true
	}
	slices.Sort(apps)
	return slices.Compact(apps), nil
}
