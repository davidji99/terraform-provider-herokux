package herokux

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func advancedDatabaseIDSchema() *schema.Schema {
	return &schema.Schema{Type: schema.TypeString, Required: true, ForceNew: true, ValidateFunc: validation.IsUUID}
}

func advancedChildID(id string) (string, string, error) {
	db, child, ok := strings.Cut(id, ":")
	if _, errs := validation.IsUUID(db, "postgres_id"); len(errs) > 0 {
		return "", "", errs[0]
	}
	if !ok || child == "" || child == "." || child == ".." {
		return "", "", fmt.Errorf("expected <postgres_id>:<resource_identifier>, got %q", id)
	}
	return db, child, nil
}
