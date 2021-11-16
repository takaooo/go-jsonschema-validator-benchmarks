# Go JSON Schema Validator Benchmarks

Adapted from [TheWildBlue](https://github.com/wbvinc/go-jsonschema-validator-benchmarks)

## Systems Under Test

http://github.com/qri-io/jsonschema

http://github.com/xeipuuv/gojsonschema

http://github.com/santhosh-tekuri/jsonschema

## Method

* Tested against [draft7](https://github.com/json-schema-org/JSON-Schema-Test-Suite/tree/master/tests/draft7) and [2019-09](https://github.com/json-schema-org/JSON-Schema-Test-Suite/tree/master/tests/draft2019-09) and except `/optional` and `refRemote.json`
* Testing for `validate()` *speed*, not correct `validate()` *result*
* Timed only canonical use of each implementation according to its own docs, not time spent converting input to the appropriate format for each

## Results for Draft 7

* All `qro-io` implementation `validate()` calls succeed with `2` incorrect results and required `3528496 ns` to test all cases 
* All `xeipuuv` implementation `validate()` calls succeed with `no` incorrect results and required `6731991 ns` to test all cases
* `santosh-tekuri` implementation failed at or before `validate()` for many cases, so I skipped the offending files. There were `no` incorrect results in cases where `validate()` succeeded, and it took `23028484 ns` to test all cases.

```	
// from BenchmarkSanthosh(), skipped all cases from these files to pass benchmark
if s.src == "definitions.json" ||
    s.src == "ref.json" {
    continue
}
```

## Results for 2019-09
* All `qro-io` implementation `validate()` calls succeed with `457` incorrect results and required `1391018428 ns` to test all cases
* All `xeipuuv` implementation `validate()` calls succeed with `107` incorrect results and required `4262408928 ns` to test all cases
* `santosh-tekuri` implementation failed at or before `validate()` for many cases, so files `ref.json` and `vocabulary.json` were skipped. There were `no` incorrect results in cases where `validate()` succeeded, and it took `107465337 ns` to test all cases.
