package advanced

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Pool struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Level  string `json:"expected_level"`
	Count  int    `json:"expected_count"`
}

func (c *Client) CreateFollowerPool(ctx context.Context, db, name, level string, count int) (Pool, error) {
	var pool Pool
	err := c.request(ctx, http.MethodPost, databasePath(db)+"/pools", map[string]any{"name": name, "level": level, "count": count}, &pool)
	return pool, err
}

// GetFollowerPool resolves a pool by its stable ID and excludes the leader from management.
func (c *Client) GetFollowerPool(ctx context.Context, db, id string) (*Pool, error) {
	var info struct {
		Pools []Pool `json:"pools"`
	}
	if err := c.request(ctx, http.MethodGet, databasePath(db)+"/info", nil, &info); err != nil {
		return nil, err
	}
	if info.Pools == nil {
		return nil, fmt.Errorf("Heroku returned database info without the pools list")
	}
	var match *Pool
	seen := make(map[string]bool, len(info.Pools))
	for i := range info.Pools {
		pool := &info.Pools[i]
		if strings.TrimSpace(pool.ID) == "" || seen[pool.ID] {
			return nil, fmt.Errorf("Heroku returned a missing or duplicate pool ID in the pools list")
		}
		seen[pool.ID] = true
		if pool.ID == id {
			if pool.Name == "" || strings.EqualFold(pool.Name, "leader") {
				return nil, fmt.Errorf("pool %q is not an identifiable follower (name %q)", id, pool.Name)
			}
			match = pool
		}
	}
	if match != nil {
		return match, nil
	}
	return nil, ErrNotFound
}

func (c *Client) ResizeFollowerPool(ctx context.Context, db, id string, change map[string]any) error {
	return c.request(ctx, http.MethodPatch, databasePath(db)+"/pools/"+url.PathEscape(id), change, nil)
}

func (c *Client) DeleteFollowerPool(ctx context.Context, db, id string) error {
	return c.request(ctx, http.MethodDelete, databasePath(db)+"/pools/"+url.PathEscape(id), nil, nil)
}
