package herokux

import (
	"context"
	"reflect"
	"testing"
)

func TestResourceHerokuxPipelineGithubIntegrationStateUpgradeV0(t *testing.T) {
	testCases := []struct {
		Description   string
		InputState    map[string]interface{}
		ExpectedState map[string]interface{}
	}{
		{
			Description:   "missing state",
			InputState:    nil,
			ExpectedState: nil,
		},
		{
			Description: "changes github_org_repo to org_repo",
			InputState: map[string]interface{}{
				"pipeline_id":     "pipeline_UUID",
				"github_org_repo": "mycompany/myrepo",
			},
			ExpectedState: map[string]interface{}{
				"pipeline_id": "pipeline_UUID",
				"org_repo":    "mycompany/myrepo",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.Description, func(t *testing.T) {
			got, err := resourceHerokuxPipelineGithubIntegrationStateUpgradeV0(context.Background(),
				testCase.InputState, nil)

			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}

			if !reflect.DeepEqual(testCase.ExpectedState, got) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", testCase.ExpectedState, got)
			}

		})
	}
}
