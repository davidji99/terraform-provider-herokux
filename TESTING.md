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

### Test Parameters

The following parameters are available for running the test. The absence of some of the non-required parameters will cause certain tests to be skipped.

* **TF_ACC** (`integer`) **Required** - must be set to `1`.
* **HEROKU_API_KEY** (`string`) **Required**  - A valid Heroku API key.
* **HEROKUX_APP_ID** (`string`) - The UUID of an existing app.
* **HEROKUX_DB_NAME** (`string`) - The name of an existing postgres database.
* **HEROKUX_KAFKA_ID** (`string`) - The UUID of an existing Kafka addon.

**For example:**
```bash
export TF_ACC=1
export HEROKU_API_KEY=...
export HEROKUX_APP_ID=...
$ make testacc TEST="./HerokuxFormationAutoscaling/" 2>&1 | tee test.log
```
