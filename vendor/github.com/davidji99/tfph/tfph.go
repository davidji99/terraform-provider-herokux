package tfph

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"
)

func ParseCompositeID(id string, numOfSplits int, separator ...string) ([]string, error) {
	sep := ":"
	if len(separator) > 0 {
		sep = separator[0]
	}

	parts := strings.Split(id, sep)

	if len(parts) != numOfSplits {
		return nil, fmt.Errorf("error: composite ID requires %d parts separated by a [%[2]s] (x%[2]sy)",
			numOfSplits, sep)
	}

	return parts, nil
}

func ContainsString(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func DoesNotContainString(s []string, str string) bool {
	return !ContainsString(s, str)
}

func ErrsFromDiags(diags diag.Diagnostics) error {
	if !diags.HasError() {
		return nil
	}

	var err string
	for _, d := range diags {
		err += fmt.Sprintf("Severity: %d | Summary: %s, | Detail: %s\n", d.Severity, d.Summary, d.Detail)
	}
	return fmt.Errorf(err)
}
