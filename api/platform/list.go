package platform

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/davidji99/simpleresty"
)

func (p *Platform) ListApps(ctx context.Context) ([]heroku.App, error) {
	apps, _, err := listPages[heroku.App](ctx, p.http, p.http.RequestURL("/apps"))
	return apps, err
}

// ListAddons lists add-ons for an app, or all accessible add-ons when app is empty.
func (p *Platform) ListAddons(ctx context.Context, app string) ([]heroku.AddOn, error) {
	path := p.http.RequestURL("/addons")
	if app != "" {
		path = p.http.RequestURL("/apps/%s/addons", url.PathEscape(app))
	}
	addons, _, err := listPages[heroku.AddOn](ctx, p.http, path)
	return addons, err
}

func listPages[T any](ctx context.Context, client *simpleresty.Client, path string) ([]T, *simpleresty.Response, error) {
	items := []T{}
	seen := map[string]bool{}
	for next := ""; ; {
		var page []T
		req := client.R().SetContext(ctx).SetResult(&page)
		req.Method, req.URL = http.MethodGet, path
		if next != "" {
			req.SetHeader("Range", next)
		}
		resp, err := client.Dispatch(req)
		// simpleresty treats 206 as an error even when it decoded the page successfully.
		if err != nil && (resp == nil || resp.StatusCode != http.StatusPartialContent) {
			return nil, resp, err
		}
		if (resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent) || page == nil {
			return nil, resp, fmt.Errorf("invalid list response from %s (HTTP %d)", path, resp.StatusCode)
		}
		items = append(items, page...)
		next = resp.Resp.Header().Get("Next-Range")
		if next == "" && resp.StatusCode == http.StatusOK {
			return items, resp, nil
		}
		if next == "" || seen[next] {
			return nil, resp, fmt.Errorf("missing or repeated pagination range from %s", path)
		}
		seen[next] = true
	}
}
