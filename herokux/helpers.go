package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
)

// getAppId extracts the app attribute generically from a HerokuX resource.
func getAppId(d *schema.ResourceData) string {
	var appName string
	if v, ok := d.GetOk("app_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] app_id name: %s", vs)
		appName = vs
	}

	return appName
}

// getFormationName extracts the formation name attribute generically from a HerokuX resource.
func getFormationName(d *schema.ResourceData) string {
	var formationName string
	if v, ok := d.GetOk("formation_name"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] formation_name: %s", vs)
		formationName = vs
	}

	return formationName
}

func parseCompositeID(id string, numOfSplits int) ([]string, error) {
	parts := strings.SplitN(id, ":", numOfSplits)

	if len(parts) != numOfSplits {
		return nil, fmt.Errorf("Error: import composite ID requires %d parts separated by a colon (x:y). "+
			"Please check resource documentation for more information.", numOfSplits)
	}
	return parts, nil
}
