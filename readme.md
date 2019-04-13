# Go JSON Schema Validator Benchmarks (Draft 7 Only)

*April 11, 2019*

[TheWildBlue](http://github.com/TheWildBlue)

## Systems Under Test

http://github.com/qri-io/jsonschema

http://github.com/xeipuuv/gojsonschema

http://github.com/santhosh-tekuri/jsonschema

## Method

* Tested against http://github.com/json-schema-org/JSON-Schema-Test-Suite/tests/draft7 except `/optional` and `refRemote.json` (I didn't feel like writing the boilerplate that would allow the remote refs to resolve)
* Testing for `validate()` *speed*, not correct `validate()` *result*
* Timed only canonical use of each implementation according to its own docs, not time spent converting input to the appropriate format for each

## Results

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

```
// from BenchmarkSanthosh(), skipped all cases from these files to pass benchmark
if s.src == "const.json" ||
    s.src == "contains.json" ||
    s.src == "enum.json" ||
    s.src == "uniqueItems.json" {
    continue
}
```
### Raw Results

```
BenchmarkQri-8        	     500	   3528496 ns/op	 1417909 B/op	   20094 allocs/op
--- BENCH: BenchmarkQri-8
    $GOPATH\src\github.com\TheWildBlue\validator-benchmarks\validator_test.go:99: incorrect validate. schema file: definitions.json. schema desc: valid definition. test desc: valid definition schema.
    $GOPATH\src\github.com\TheWildBlue\validator-benchmarks\validator_test.go:99: incorrect validate. schema file: ref.json. schema desc: remote ref, containing refs itself. test desc: remote ref valid.
	... [output truncated]
BenchmarkXeipuu-8     	     200	   6731991 ns/op	 4643779 B/op	   63657 allocs/op
BenchmarkSanthosh-8   	      50	  23028484 ns/op	 2527688 B/op	   23763 allocs/op
PASS
ok  	github.com/TheWildBlue/validator-benchmarks	16.058s
Success: Benchmarks passed.
```

## Conclusion

`qri-io` implementation is generally ~2x faster than `xeipuu`. 

`xeipuu` is generally ~3.5x than `santhosh-tekuri`, even with many cases skipped for the latter (thus `qri-io` is generally ~7x faster than `santhosh-tekuri`).

`qri-io` does not validate `2` cases from `definitions.json` and `ref.json` correctly.

`santhosh-tekuri` failed for many cases.
