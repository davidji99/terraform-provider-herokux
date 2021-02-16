package herokux

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// getAppID extracts the app ID attribute generically from a HerokuX resource.
func getAppID(d *schema.ResourceData) string {
	var appID string
	if v, ok := d.GetOk("app_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] app_id: %s", vs)
		appID = vs
	}

	return appID
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

// getProcessType extracts the process type attribute generically from a HerokuX resource.
func getProcessType(d *schema.ResourceData) string {
	var processType string
	if v, ok := d.GetOk("process_type"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] process_type: %s", vs)
		processType = vs
	}

	return processType
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

// getKafkaID extracts the kafka/cluster ID attribute generically from a HerokuX resource.
func getKafkaID(d *schema.ResourceData) string {
	var kafkaID string
	if v, ok := d.GetOk("kafka_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] kafka_id: %s", vs)
		kafkaID = vs
	}

	return kafkaID
}

// getPostgresID extracts the postgres ID attribute generically from a HerokuX resource.
func getPostgresID(d *schema.ResourceData) string {
	var postgresID string
	if v, ok := d.GetOk("postgres_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] postgres_id: %s", vs)
		postgresID = vs
	}

	return postgresID
}

// getRedisID extracts the redis ID attribute generically from a HerokuX resource.
func getRedisID(d *schema.ResourceData) string {
	var redisID string
	if v, ok := d.GetOk("redis_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] redis_id: %s", vs)
		redisID = vs
	}

	return redisID
}

// getConnectID extracts the connect ID attribute generically from a HerokuX resource.
func getConnectID(d *schema.ResourceData) string {
	var connectID string
	if v, ok := d.GetOk("connect_id"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] connect_id: %s", vs)
		connectID = vs
	}

	return connectID
}

func getConnectMappings(d *schema.ResourceData) []byte {
	var mappings []byte
	if v, ok := d.GetOk("mappings"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] mappings: %s", vs)
		mappings = []byte(vs)
	}

	return mappings
}

func parseCompositeID(id string, numOfSplits int) ([]string, error) {
	parts := strings.SplitN(id, ":", numOfSplits)

	if len(parts) != numOfSplits {
		return nil, fmt.Errorf("Error: import composite ID requires %d parts separated by a colon (x:y). "+
			"Please check resource documentation for more information.", numOfSplits)
	}
	return parts, nil
}

func parseCompositeIDCustom(id, sep string, numOfSplits int) ([]string, error) {
	parts := strings.SplitN(id, sep, numOfSplits)

	if len(parts) != numOfSplits {
		return nil, fmt.Errorf("Error: import composite ID requires %d parts each separated by a `%s`. "+
			"Please check resource documentation for more information.", numOfSplits, sep)
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

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func getRandomStringFromSlice(s []string) string {
	index := randInt(0, len(s)-1)
	return s[index]
}

func ContainsString(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func validateMaintenanceWindow(v interface{}, k string) (ws []string, errors []error) {
	name := v.(string)
	if !regexp.MustCompile(`^[A-Za-z]{2,10}s \d\d?:[03]0$`).MatchString(name) {
		errors = append(errors, fmt.Errorf("maintenance window format should be 'Days HH:MM' where where MM is 00 or 30"))
	}
	return
}
