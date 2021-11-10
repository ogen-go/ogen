// Package ogen implements OpenAPI v3 code generation.
package ogen

//go:generate go run ./cmd/ogen --schema _testdata/sample.json --target internal/sample_api --clean
//go:generate go run ./cmd/ogen --schema _testdata/errors.json --target internal/sample_err --clean
//go:generate go run ./cmd/ogen --schema _testdata/techempower.json --target internal/techempower --package techempower --clean
