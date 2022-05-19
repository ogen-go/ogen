package examples

import (
	_ "github.com/go-faster/errors"

	_ "github.com/ogen-go/ogen"
)

// Generate schemas:

//go:generate go run github.com/ogen-go/ogen/tools/mkformattest --output ../_testdata/positive/test_format.json

// Generate JSON Schema:

//go:generate go run github.com/ogen-go/ogen/cmd/jschemagen --typename Spec --target ex_openapi/output.gen.go ../gen/_testdata/jsonschema/openapi30.json

// Fully supported:

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_test_format --clean --generate-tests ../_testdata/positive/test_format.json

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_petstore --clean --generate-tests  ../_testdata/positive/petstore.yml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_firecracker --clean --generate-tests  ../_testdata/positive/firecracker.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_gotd --clean --generate-tests  ../_testdata/positive/gotd_bot_api.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_ent --clean --generate-tests  ../_testdata/positive/ent.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_route_params --clean --generate-tests  ../_testdata/positive/ex_route_params.json

// Partially supported:

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_manga --clean --debug.ignoreNotImplemented "unsupported content types" --generate-tests  ../_testdata/positive/manga.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_petstore_expanded --clean --debug.noerr --generate-tests  ../_testdata/positive/petstore-expanded.yml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_k8s --clean --debug.noerr --generate-tests  ../_testdata/positive/k8s.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_github --clean --infer-types --debug.noerr --generate-tests  ../_testdata/positive/api.github.com.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_telegram --clean --debug.noerr --generate-tests  ../_testdata/positive/telegram_bot_api.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_tinkoff --clean --debug.ignoreNotImplemented "http security" --generate-tests  ../_testdata/positive/tinkoff.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target ex_2ch --clean --debug.ignoreNotImplemented "unsupported content types" --generate-tests  ../_testdata/positive/2ch.yml
