package integration

// Sample API matrix:
//
//go:generate go run ../../cmd/ogen -v --clean --config _config/sample_api.yml         --target sample_api         ../../_testdata/positive/sample.json
//go:generate go run ../../cmd/ogen -v --clean --config _config/sample_api_ns.yml      --target sample_api_ns      ../../_testdata/positive/sample.json
//go:generate go run ../../cmd/ogen -v --clean --config _config/sample_api_nc.yml      --target sample_api_nc      ../../_testdata/positive/sample.json
//go:generate go run ../../cmd/ogen -v --clean --config _config/sample_api_nsnc.yml    --target sample_api_nsnc    ../../_testdata/positive/sample.json
//go:generate go run ../../cmd/ogen -v --clean --config _config/sample_api_no_otel.yml --target sample_api_no_otel ../../_testdata/positive/sample.json

//go:generate go run ../../cmd/ogen -v --clean --config _config/errors.yml --target sample_err ../../_testdata/positive/convenient_errors/errors.json
//go:generate go run ../../cmd/ogen -v --clean --config _config/techempower.yml --package techempower --target techempower ../../_testdata/examples/techempower.json

//go:generate go run ../../cmd/ogen -v --clean --config _config/security_reentrant.yml --target security_reentrant ../../_testdata/positive/security.json

// Tests
//
//go:generate go run ../../cmd/ogen -v --clean --target test_webhooks         ../../_testdata/positive/webhooks.json
//go:generate go run ../../cmd/ogen -v --clean --target test_servers          ../../_testdata/positive/servers.json
//go:generate go run ../../cmd/ogen -v --clean --target test_single_endpoint  ../../_testdata/positive/single_endpoint.json
//go:generate go run ../../cmd/ogen -v --clean --target test_http_responses   ../../_testdata/positive/http_responses.json
//go:generate go run ../../cmd/ogen -v --clean --target test_http_requests    ../../_testdata/positive/http_requests.json
//go:generate go run ../../cmd/ogen -v --clean --target test_form             ../../_testdata/positive/form.json
//go:generate go run ../../cmd/ogen -v --clean --target test_parameters       ../../_testdata/positive/parameters.json
//go:generate go run ../../cmd/ogen -v --clean --target test_security         ../../_testdata/positive/security.json
//
//
//go:generate go run ../../cmd/ogen -v --clean --target referenced_path_item ../../_testdata/positive/referenced_pathItem.json
//
//go:generate go run ../../cmd/ogen -v --clean --config _config/allOf.yml --target test_allof ../../_testdata/positive/allOf.yml
//go:generate go run ../../cmd/ogen -v --clean --config _config/anyOf.yml --target test_anyof ../../_testdata/positive/anyOf.json
//go:generate go run ../../cmd/ogen -v --clean --target test_discriminator_mapping ../../_testdata/positive/discriminator_mapping.json
//go:generate go run ../../cmd/ogen -v --clean --config _config/additionalPropertiesPatternProperties.yml --target test_additionalpropertiespatternproperties ../../_testdata/positive/additionalPropertiesPatternProperties.yml
//go:generate go run ../../cmd/ogen -v --clean --config _config/client_options.yml --target test_client_options ../../_testdata/positive/client_options.json
//
//go:generate go run ../../cmd/ogen -v --clean -target test_enum_naming       ../../_testdata/positive/enum_naming.yml
//go:generate go run ../../cmd/ogen -v --clean -target test_naming_extensions ../../_testdata/positive/naming_extensions.json
//go:generate go run ../../cmd/ogen -v --clean -target test_type_extension ../../_testdata/positive/type_extension.yml
//go:generate go run ../../cmd/ogen -v --clean -target test_type_extension_name ../../_testdata/positive/type_extension_name.yml
//go:generate go run ../../cmd/ogen -v --clean -target test_time_extension ../../_testdata/positive/time_extension.yml
//
// Regression test.
//
//go:generate go run ../../cmd/ogen -v --clean --target test_issue1161 ../../_testdata/positive/issue1161.json
