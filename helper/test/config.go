package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type TestConfigKey int

const (
	TestConfigHerokuxAPIKey TestConfigKey = iota
	TestConfigHerokuxCustomAPIKey
	TestConfigAddonID
	TestConfigAppID
	TestConfigDatabaseName
	TestConfigKafkaID
	TestConfigTeamID
	TestConfigRedisID
	TestConfigPostgresID
	TestConfigConnectID
	TestConfigImageID
	TestConfigPipelineID
	TestConfigGithubOrgRepo
	TestConfigUserEmail
	TestConfigOrganization
	TestConfigAcceptanceTestKey
)

var testConfigKeyToEnvName = map[TestConfigKey]string{
	TestConfigHerokuxAPIKey:       "HEROKU_API_KEY",
	TestConfigHerokuxCustomAPIKey: "HEROKUX_TESTACC_API_KEY",
	TestConfigAddonID:             "HEROKUX_ADDON_ID",
	TestConfigAppID:               "HEROKUX_APP_ID",
	TestConfigDatabaseName:        "HEROKUX_DB_NAME",
	TestConfigKafkaID:             "HEROKUX_KAFKA_ID",
	TestConfigTeamID:              "HEROKUX_TEAM_ID",
	TestConfigRedisID:             "HEROKUX_REDIS_ID",
	TestConfigPostgresID:          "HEROKUX_POSTGRES_ID",
	TestConfigConnectID:           "HEROKUX_CONNECT_ID",
	TestConfigImageID:             "HEROKUX_IMAGE_ID",
	TestConfigPipelineID:          "HEROKUX_PIPELINE_ID",
	TestConfigGithubOrgRepo:       "HEROKUX_GITHUB_ORG_REPO",
	TestConfigUserEmail:           "HEROKUX_USER_EMAIL",
	TestConfigOrganization:        "HEROKUX_ORGANIZATION",
	TestConfigAcceptanceTestKey:   resource.TestEnvVar,
}

func (k TestConfigKey) String() (name string) {
	if val, ok := testConfigKeyToEnvName[k]; ok {
		name = val
	}
	return
}

type TestConfig struct{}

func NewTestConfig() *TestConfig {
	return &TestConfig{}
}

func (t *TestConfig) Get(keys ...TestConfigKey) (val string) {
	for _, key := range keys {
		val = os.Getenv(key.String())
		if val != "" {
			break
		}
	}
	return
}

func (t *TestConfig) GetOrSkip(testing *testing.T, keys ...TestConfigKey) (val string) {
	t.SkipUnlessAccTest(testing)
	val = t.Get(keys...)
	if val == "" {
		testing.Skip(fmt.Sprintf("skipping test: config %v not set", keys))
	}
	return
}

func (t *TestConfig) GetOrAbort(testing *testing.T, keys ...TestConfigKey) (val string) {
	t.SkipUnlessAccTest(testing)
	val = t.Get(keys...)
	if val == "" {
		testing.Fatal(fmt.Sprintf("stopping test: config %v must be set", keys))
	}
	return
}

func (t *TestConfig) SkipUnlessAccTest(testing *testing.T) {
	val := t.Get(TestConfigAcceptanceTestKey)
	if val == "" {
		testing.Skip(fmt.Sprintf("Acceptance tests skipped unless env '%s' set", TestConfigAcceptanceTestKey.String()))
	}
}

func (t *TestConfig) GetAddonIDorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigAddonID)
}

func (t *TestConfig) GetAppIDorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigAppID)
}

func (t *TestConfig) GetDBNameorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigDatabaseName)
}

func (t *TestConfig) GetKafkaIDorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigKafkaID)
}

func (t *TestConfig) GetCustomAPIKeyorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigHerokuxCustomAPIKey)
}

func (t *TestConfig) GetTeamIDorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigTeamID)
}

func (t *TestConfig) GetRedisIDorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigRedisID)
}

func (t *TestConfig) GetPostgresIDorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigPostgresID)
}

func (t *TestConfig) GetConnectIDorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigConnectID)
}

func (t *TestConfig) GetImageIDorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigImageID)
}

func (t *TestConfig) GetPipelineIDorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigPipelineID)
}

func (t *TestConfig) GetGithubOrgRepoorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigGithubOrgRepo)
}

func (t *TestConfig) GetUserEmailorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigUserEmail)
}

func (t *TestConfig) GetAnyOrganizationOrSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigOrganization)
}
