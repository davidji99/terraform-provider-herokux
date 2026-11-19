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
	for _, pool := range info.Pools {
		if pool.ID == id {
			if pool.Name == "" || strings.EqualFold(pool.Name, "leader") {
				return nil, fmt.Errorf("pool %q is not an identifiable follower (name %q)", id, pool.Name)
			}
			return &pool, nil
		}
	}
	return nil, ErrNotFound
}

func (c *Client) ResizeFollowerPool(ctx context.Context, db, id string, change map[string]any) error {
	return c.request(ctx, http.MethodPatch, databasePath(db)+"/pools/"+url.PathEscape(id), change, nil)
}

func (c *Client) DeleteFollowerPool(ctx context.Context, db, id string) error {
	return c.request(ctx, http.MethodDelete, databasePath(db)+"/pools/"+url.PathEscape(id), nil, nil)
}
