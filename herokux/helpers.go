package herokux

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
)

// getAppId extracts the app ID attribute generically from a HerokuX resource.
func getAppId(d *schema.ResourceData) string {
	var appName string
	if v, ok := d.GetOk("app_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] app_id: %s", vs)
		appName = vs
	}

	return appName
}

// getAddonID extracts the addon ID attribute generically from a HerokuX resource.
func getAddonID(d *schema.ResourceData) string {
	var addonID string
	if v, ok := d.GetOk("addon_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] addon_id: %s", vs)
		addonID = vs
	}

	return addonID
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

// getDatabaseName extracts the database name name attribute generically from a HerokuX resource.
func getDatabaseName(d *schema.ResourceData) string {
	var dbName string
	if v, ok := d.GetOk("database_name"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] database_name: %s", vs)
		dbName = vs
	}

	return dbName
}

// getKakfaID extracts the kafka/cluster ID attribute generically from a HerokuX resource.
func getKakfaID(d *schema.ResourceData) string {
	var kafkaID string
	if v, ok := d.GetOk("kafka_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] kafka_id: %s", vs)
		kafkaID = vs
	}

	return kafkaID
}

func parseCompositeID(id string, numOfSplits int) ([]string, error) {
	parts := strings.SplitN(id, ":", numOfSplits)

	if len(parts) != numOfSplits {
		return nil, fmt.Errorf("Error: import composite ID requires %d parts separated by a colon (x:y). "+
			"Please check resource documentation for more information.", numOfSplits)
	}
	return parts, nil
}

func stringArrayContains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
