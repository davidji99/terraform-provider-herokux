package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type TestConfigKey int

const (
	TestConfigHerokuplusAPIKey TestConfigKey = iota
	TestConfigAppID
	TestConfigAcceptanceTestKey
)

var testConfigKeyToEnvName = map[TestConfigKey]string{
	TestConfigHerokuplusAPIKey:  "HEROKU_API_KEY",
	TestConfigAppID:             "HEROKUPLUS_APP_ID",
	TestConfigAcceptanceTestKey: resource.TestEnvVar,
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

func (t *TestConfig) GetAppIDorSkip(testing *testing.T) (val string) {
	return t.GetOrSkip(testing, TestConfigAppID)
}
