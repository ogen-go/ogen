package internal

// Sample API matrix:
//go:generate go run ../cmd/ogen -v --target sample_api --debug.ignoreNotImplemented "enum format" --infer-types --clean --generate-tests ../_testdata/positive/sample.json
//go:generate go run ../cmd/ogen -v --no-server --target sample_api_ns --debug.ignoreNotImplemented "enum format" --infer-types --clean ../_testdata/positive/sample.json
//go:generate go run ../cmd/ogen -v --no-client --target sample_api_nc --debug.ignoreNotImplemented "enum format" --infer-types --clean ../_testdata/positive/sample.json
//go:generate go run ../cmd/ogen -v --no-server --no-client --target sample_api_nsnc --debug.ignoreNotImplemented "enum format" --infer-types --clean ../_testdata/positive/sample.json

//go:generate go run ../cmd/ogen -v --target test_single_endpoint --clean --generate-tests ../_testdata/positive/test_single_endpoint.json
//go:generate go run ../cmd/ogen -v --target test_http_responses --clean --generate-tests ../_testdata/positive/test_http_responses.json
//go:generate go run ../cmd/ogen -v --target test_http_requests --clean --generate-tests ../_testdata/positive/test_http_requests.json

//go:generate go run ../cmd/ogen -v --target sample_err --clean --generate-tests ../_testdata/positive/errors.json
//go:generate go run ../cmd/ogen -v --target techempower --package techempower --clean --generate-tests ../_testdata/positive/techempower.json

//go:generate go run ../cmd/ogen -v --target test_allof --clean --generate-tests ../_testdata/positive/test_allof.yml
