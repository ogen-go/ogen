package internal

// Sample API matrix:
//go:generate go run ../cmd/ogen -v --target sample_api --debug.ignoreNotImplemented "enum format" --infer-types --clean --generate-tests ../_testdata/positive/sample.json
//go:generate go run ../cmd/ogen -v --no-server --target sample_api_ns --debug.ignoreNotImplemented "enum format" --infer-types --clean ../_testdata/positive/sample.json
//go:generate go run ../cmd/ogen -v --no-client --target sample_api_nc --debug.ignoreNotImplemented "enum format" --infer-types --clean ../_testdata/positive/sample.json
//go:generate go run ../cmd/ogen -v --no-server --no-client --target sample_api_nsnc --debug.ignoreNotImplemented "enum format" --infer-types --clean ../_testdata/positive/sample.json

//go:generate go run ../cmd/ogen -v --target sample_err --clean ../_testdata/positive/errors.json
//go:generate go run ../cmd/ogen -v --target techempower --package techempower --clean --generate-tests ../_testdata/examples/techempower.json

// Tests
//go:generate go run ../cmd/ogen -v --target test_single_endpoint --clean ../_testdata/positive/single_endpoint.json
//go:generate go run ../cmd/ogen -v --target test_http_responses --clean ../_testdata/positive/http_responses.json
//go:generate go run ../cmd/ogen -v --target test_http_requests --clean ../_testdata/positive/http_requests.json
//go:generate go run ../cmd/ogen -v --target test_form --clean ../_testdata/positive/form.json
//go:generate go run ../cmd/ogen -v --target test_allof --clean --generate-tests ../_testdata/positive/allof.yml
