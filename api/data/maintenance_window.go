package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GetMaintenanceWindow returns the UTC range reported for the add-on.
func (d *Data) GetMaintenanceWindow(ctx context.Context, addonID string) (string, error) {
	var result struct {
		Window string `json:"window"`
	}
	err := d.maintenanceWindowRequest(ctx, http.MethodGet, addonID, nil, &result)
	return result.Window, err
}

func (d *Data) SetMaintenanceWindow(ctx context.Context, addonID, day, clock string) error {
	body := map[string]string{"day_of_week": day, "time_of_day": clock}
	return d.maintenanceWindowRequest(ctx, http.MethodPost, addonID, body, nil)
}

func (d *Data) maintenanceWindowRequest(ctx context.Context, method, addonID string, body map[string]string, result any) error {
	// Maintenance uses the same Data REST origin as Postgres Advanced. DataBaseURL serves GraphQL.
	endpoint := d.config.PostgresAdvancedBaseURL + "/data/maintenances/v1/" + url.PathEscape(addonID) + "/window"
	req := d.http.R().SetContext(ctx)
	if body != nil {
		req.SetBody(body)
	}
	resp, err := req.Execute(method, endpoint)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		var detail struct{ Message string }
		if json.Unmarshal(resp.Body(), &detail) == nil && detail.Message != "" {
			return fmt.Errorf("%s maintenance window for %s: %s: %s", method, addonID, resp.Status(), detail.Message)
		}
		return fmt.Errorf("%s maintenance window for %s: %s", method, addonID, resp.Status())
	}
	if result != nil {
		return json.Unmarshal(resp.Body(), result)
	}
	return nil
}
