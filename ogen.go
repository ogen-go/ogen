// Package ogen implements OpenAPI v3 code generation.
package ogen

//go:generate go run ./cmd/ogen --target internal/sample_api --debug.ignoreNotImplemented "enum format" --infer-types --clean --generate-tests _testdata/positive/sample.json
//go:generate go run ./cmd/ogen --target internal/test_single_endpoint --clean --generate-tests _testdata/positive/test_single_endpoint.json
//go:generate go run ./cmd/ogen --target internal/test_http_responses --clean --generate-tests _testdata/positive/test_http_responses.json

//go:generate go run ./cmd/ogen --target internal/sample_err --clean --generate-tests _testdata/positive/errors.json
//go:generate go run ./cmd/ogen --target internal/techempower --package techempower --clean --generate-tests _testdata/positive/techempower.json
