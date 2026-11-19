package advanced

import (
	"context"
	"fmt"
	"net/http"
)

type StorageQuota struct {
	WarningGB         *int   `json:"warning_gb"`
	CriticalGB        *int   `json:"critical_gb"`
	EnforcementAction string `json:"enforcement_action"`
}

func (c *Client) GetStorageQuota(ctx context.Context, db string) (StorageQuota, error) {
	var quota StorageQuota
	err := c.request(ctx, http.MethodGet, databasePath(db)+"/quotas/storage", nil, &quota)
	if err == nil && quota.EnforcementAction == "" {
		return quota, fmt.Errorf("Heroku returned a storage quota without an enforcement action")
	}
	return quota, err
}

func equalThreshold(a, b *int) bool {
	return a == b || (a != nil && b != nil && *a == *b)
}

func (c *Client) SetStorageQuota(ctx context.Context, db string, desired StorageQuota) error {
	current, err := c.GetStorageQuota(ctx, db)
	if err != nil {
		return err
	}
	changes := make(map[string]any)
	if !equalThreshold(current.WarningGB, desired.WarningGB) {
		changes["warning_gb"] = desired.WarningGB
	}
	if !equalThreshold(current.CriticalGB, desired.CriticalGB) {
		changes["critical_gb"] = desired.CriticalGB
	}
	if current.EnforcementAction != desired.EnforcementAction {
		changes["enforcement_action"] = desired.EnforcementAction
	}
	if len(changes) == 0 {
		return nil
	}
	return c.request(ctx, http.MethodPatch, databasePath(db)+"/quotas/storage", changes, nil)
}
