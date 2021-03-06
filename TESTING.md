# Testing

## Provider Tests
In order to test the provider, you can simply run `make test`.

```bash
$ make test
```

## Acceptance Tests

You can run the complete suite of HerokuX acceptance tests by doing the following:

```bash
$ make testacc TEST="./herokux/" 2>&1 | tee test.log
```

To run a single acceptance test in isolation replace the last line above with:

```bash
$ make testacc TEST="./herokux/" TESTARGS='-run=TestAccHerokuxFormationAutoscaling_Basic'
```

A set of tests can be selected by passing `TESTARGS` a substring. For example, to run all HerokuX formation autoscaling tests:

```bash
$ make testacc TEST="./herokux/" TESTARGS='-run=HerokuxFormationAutoscaling'
```

## E2E Acceptance Tests
Some, not all, resources have tests that utilize the Heroku Terraform provider for a full end-to-end testing experience.
These tests, tagged with `TestAccE2E...`, will take much longer to run and be run by setting the following environment
variable:

```shell
HEROKUX_RUN_E2E_TESTS=true
```

### Test Parameters

The following parameters are available for running the test. The absence of some of the non-required parameters will cause certain tests to be skipped.

* **TF_ACC** (`integer`) **Required** - must be set to `1`.
* **HEROKU_API_KEY** (`string`) **Required**  - A valid Heroku API key.
* **HEROKUX_TESTACC_API_KEY** (`string`)  - A valid Heroku API key. This is used for the oauth authorization acceptance tests.
* **HEROKUX_APP_ID** (`string`) - The UUID of an existing app.
* **HEROKUX_ADDON_ID** (`string`) - The UUID of an existing addon. Use this for postgres addon IDs.
* **HEROKUX_DB_NAME** (`string`) - The name of an existing postgres database.
* **HEROKUX_KAFKA_ID** (`string`) - The UUID of an existing Kafka addon.
* **HEROKUX_REDIS_ID** (`string`) - The UUID of an existing Redis addon.
* **HEROKUX_POSTGRES_ID** (`string`) - The UUID of an existing Postgres addon.
* **HEROKUX_CONNECT_ID** (`string`) - The UUID of an existing Heroku Connect integration.
* **HEROKUX_IMAGE_ID** (`string`) - The UUID of an existing docker image in Heroku.
* **HEROKUX_PIPELINE_ID** (`string`) - The UUID of an existing pipeline.
* **HEROKUX_GITHUB_ORG_REPO** (`string`) - The org/repo of an existing Github repository.
* **HEROKUX_USER_EMAIL** (`string`) - Email address of an existing Heroku user.
* **HEROKUX_RUN_E2E_TESTS** (`string`) - Execute integration tests that make use of the Heroku provider for a true E2E
  test experience. These tests may take a long time as they create all required upstream resources. Set to `"true"` to
  execute these tests.

**For example:**
```bash
export TF_ACC=1
export HEROKU_API_KEY=...
export HEROKUX_APP_ID=...
$ make testacc TEST="./HerokuxFormationAutoscaling/" 2>&1 | tee test.log
```
