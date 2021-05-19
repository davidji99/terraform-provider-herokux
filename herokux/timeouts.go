package herokux

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"strconv"
	"time"
)

// resourceTimeouts returns predefined timeouts applicable for resources, not data sources.
//
// These timeouts have default values and can be overridden via their equivalent environment variable.
func resourceTimeouts() *schema.ResourceTimeout {
	resourceCreateTimeout := 90
	resourceReadTimeout := 10
	resourceUpdateTimeout := 60
	resourceDeleteTimeout := 30

	if v, ok := os.LookupEnv("HEROKUX_RESOURCE_GLOBAL_CREATE_TIMEOUT"); ok {
		resourceCreateTimeout, _ = strconv.Atoi(v)
	}

	if v, ok := os.LookupEnv("HEROKUX_RESOURCE_GLOBAL_READ_TIMEOUT"); ok {
		resourceReadTimeout, _ = strconv.Atoi(v)
	}

	if v, ok := os.LookupEnv("HEROKUX_RESOURCE_GLOBAL_UPDATE_TIMEOUT"); ok {
		resourceUpdateTimeout, _ = strconv.Atoi(v)
	}

	if v, ok := os.LookupEnv("HEROKUX_RESOURCE_GLOBAL_DELETE_TIMEOUT"); ok {
		resourceDeleteTimeout, _ = strconv.Atoi(v)
	}

	return &schema.ResourceTimeout{
		Create:  schema.DefaultTimeout(time.Duration(resourceCreateTimeout) * time.Minute),
		Read:    schema.DefaultTimeout(time.Duration(resourceReadTimeout) * time.Minute),
		Update:  schema.DefaultTimeout(time.Duration(resourceUpdateTimeout) * time.Minute),
		Delete:  schema.DefaultTimeout(time.Duration(resourceDeleteTimeout) * time.Minute),
		Default: schema.DefaultTimeout(45 * time.Minute),
	}
}
