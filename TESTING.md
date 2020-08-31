# Testing

## Provider Tests
In order to test the provider, you can simply run `make test`.

```bash
$ make test
```

## Acceptance Tests

You can run the complete suite of Herokuplus acceptance tests by doing the following:

```bash
$ make testacc TEST="./herokux/" 2>&1 | tee test.log
```

To run a single acceptance test in isolation replace the last line above with:

```bash
$ make testacc TEST="./herokux/" TESTARGS='-run=TestAccHerokuplusFormationAutoscaling_Basic'
```

A set of tests can be selected by passing `TESTARGS` a substring. For example, to run all Herokuplus formation autoscaling tests:

```bash
$ make testacc TEST="./herokux/" TESTARGS='-run=HerokuplusFormationAutoscaling'
```

### Test Parameters

The following parameters are available for running the test. The absence of some of the non-required parameters will cause certain tests to be skipped.

* **TF_ACC** (`integer`) **Required** - must be set to `1`.
* **HEROKU_API_KEY** (`string`) **Required**  - A valid Heroku API key.
* **HEROKUPLUS_APP_ID** (`string`) - THe UUID of an existing app.

**For example:**
```bash
export TF_ACC=1
export HEROKU_API_KEY=...
export HEROKUPLUS_APP_ID=...
$ make testacc TEST="./HerokuplusFormationAutoscaling/" 2>&1 | tee test.log
```
