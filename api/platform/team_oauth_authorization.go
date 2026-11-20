package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	heroku "github.com/davidji99/heroku-go/v5"
	"github.com/hashicorp/go-uuid"
)

func (p *Platform) CreateTeamOAuthAuthorization(ctx context.Context, team string, opts heroku.OAuthAuthorizationCreateOpts) (*heroku.OAuthAuthorization, error) {
	path := p.teamOAuth.RequestURL("/teams/%s/oauth/authorizations", url.PathEscape(team))
	resp, err := p.teamOAuth.R().SetContext(ctx).SetBody(opts).Post(path)
	if err != nil {
		return nil, err
	}
	// The response body can contain a bearer token, including on errors.
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("creating OAuth authorization for team %s: HTTP %d", team, resp.StatusCode())
	}
	var auth heroku.OAuthAuthorization
	if json.Unmarshal(resp.Body(), &auth) != nil {
		return nil, fmt.Errorf("invalid create response for team %s OAuth authorization", team)
	}
	if _, err := uuid.ParseUUID(auth.ID); err != nil {
		return nil, fmt.Errorf("missing or invalid OAuth authorization ID in create response for team %s", team)
	}
	return &auth, nil
}

// FindTeamOAuthAuthorization returns nil only after a complete team listing excludes id.
func (p *Platform) FindTeamOAuthAuthorization(ctx context.Context, team, id string) (*heroku.OAuthAuthorization, error) {
	path := p.teamOAuth.RequestURL("/teams/%s/oauth/authorizations", url.PathEscape(team))
	seen := map[string]bool{}
	for next := ""; ; {
		req := p.teamOAuth.R().SetContext(ctx)
		if next != "" {
			req.SetHeader("Range", next)
		}
		resp, err := req.Get(path)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusPartialContent {
			return nil, fmt.Errorf("listing OAuth authorizations for team %s: HTTP %d", team, resp.StatusCode())
		}
		var page []heroku.OAuthAuthorization
		if json.Unmarshal(resp.Body(), &page) != nil || page == nil {
			return nil, fmt.Errorf("invalid OAuth authorization list for team %s", team)
		}
		for _, auth := range page {
			if auth.ID == "" {
				return nil, fmt.Errorf("missing authorization ID in list for team %s", team)
			}
			if strings.EqualFold(auth.ID, id) {
				if auth.Client != nil {
					return nil, fmt.Errorf("authorization %s belongs to an OAuth client (expected a direct team authorization)", id)
				}
				return &auth, nil
			}
		}
		next = resp.Header().Get("Next-Range")
		if next == "" && resp.StatusCode() == http.StatusOK {
			return nil, nil
		}
		if next == "" || seen[next] {
			return nil, fmt.Errorf("missing or repeated OAuth authorization pagination range for team %s", team)
		}
		seen[next] = true
	}
}
