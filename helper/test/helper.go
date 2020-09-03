package test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func GenerateRandomCIDR() string {
	return fmt.Sprintf("%d.%d.%d.%d/32", acctest.RandIntRange(1, 10), acctest.RandIntRange(1, 10),
		acctest.RandIntRange(1, 10), acctest.RandIntRange(1, 10))
}
