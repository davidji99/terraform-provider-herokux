package tfph

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func GetStringValue(d *schema.ResourceData, valueName string) string {
	var x string
	if v, ok := d.GetOk(valueName); ok {
		x = v.(string)
		log.Printf("[DEBUG] %s: %s", valueName, x)
	}

	return x
}

func GetIntValue(d *schema.ResourceData, valueName string) int {
	var i int
	if v, ok := d.GetOk(valueName); ok {
		i = v.(int)
		log.Printf("[DEBUG] %s: %d", valueName, i)
	}

	return i
}

func GetListValueAsStringSlice(d *schema.ResourceData, valueName string) []string {
	var x []string
	if v, ok := d.GetOk(valueName); ok {
		for _, i := range v.([]interface{}) {
			x = append(x, i.(string))
		}
		log.Printf("[DEBUG] %s: %v", valueName, x)
	}

	return x
}

func GetListValueAsIntSlice(d *schema.ResourceData, valueName string) []int {
	var x []int
	if v, ok := d.GetOk(valueName); ok {
		for _, i := range v.([]interface{}) {
			x = append(x, i.(int))
		}
		log.Printf("[DEBUG] %s: %v", valueName, x)
	}

	return x
}

func GetSetValueAsStringSlice(d *schema.ResourceData, valueName string) []string {
	var x []string
	if v, ok := d.GetOk(valueName); ok {
		for _, i := range v.(*schema.Set).List() {
			x = append(x, i.(string))
		}
		log.Printf("[DEBUG] %s: %v", valueName, x)
	}

	return x
}

func GetSetValueAsIntSlice(d *schema.ResourceData, valueName string) []int {
	var x []int
	if v, ok := d.GetOk(valueName); ok {
		for _, i := range v.(*schema.Set).List() {
			x = append(x, i.(int))
		}
		log.Printf("[DEBUG] %s: %v", valueName, x)
	}

	return x
}
