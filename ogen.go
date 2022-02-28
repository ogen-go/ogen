// Package ogen implements OpenAPI v3 code generation.
package ogen

//go:generate go run ./cmd/ogen --schema _testdata/positive/sample.json --target internal/sample_api --infer-types --clean --generate-tests
//go:generate go run ./cmd/ogen --schema _testdata/positive/test_single_endpoint.json --target internal/test_single_endpoint  --clean --generate-tests
//go:generate go run ./cmd/ogen --schema _testdata/positive/test_http_responses.json --target internal/test_http_responses  --clean --generate-tests

//go:generate go run ./cmd/ogen --schema _testdata/positive/errors.json --target internal/sample_err --clean --generate-tests
//go:generate go run ./cmd/ogen --schema _testdata/positive/techempower.json --target internal/techempower --package techempower --clean --generate-tests
