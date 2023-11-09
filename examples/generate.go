package examples

import (
	_ "github.com/ogen-go/ogen"
)

// Generate schemas:
//
//go:generate go run github.com/ogen-go/ogen/tools/mkformattest --output ../_testdata/positive/format_gen.json

// Generate JSON Schema:
//
//go:generate go run github.com/ogen-go/ogen/cmd/jschemagen --typename Spec --target ex_openapi/output.gen.go ../gen/_testdata/jsonschema/openapi30.json

// Fully supported:
//
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/with_tests.yml --target ex_test_format       ../_testdata/positive/format_gen.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/with_tests.yml --target ex_petstore          ../_testdata/examples/petstore.yml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/with_tests.yml --target ex_firecracker       ../_testdata/examples/firecracker.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/with_tests.yml --target ex_gotd              ../_testdata/examples/gotd_bot_api.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/with_tests.yml --target ex_ent               ../_testdata/examples/ent.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/with_tests.yml --target ex_route_params      ../_testdata/positive/ex_route_params.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/with_tests.yml --target ex_manga             ../_testdata/examples/manga.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/with_tests.yml --target ex_petstore_expanded ../_testdata/examples/petstore-expanded.yml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/with_tests.yml --target ex_telegram          ../_testdata/examples/telegram_bot_api.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/with_tests.yml --target ex_2ch               ../_testdata/examples/2ch.yml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/with_tests.yml --target ex_tinkoff           ../_testdata/examples/tinkoff.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --target ex_openai                                            ../_testdata/examples/openai.yml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --target ex_oauth2                                            ../_testdata/examples/petstore-oauth2.yml

// Partially supported:
//
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/k8s.yml    --target ex_k8s    ../_testdata/examples/k8s.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --config _config/github.yml --target ex_github ../_testdata/examples/api.github.com.json
