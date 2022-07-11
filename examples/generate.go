package examples

import (
	_ "github.com/go-faster/errors"

	_ "github.com/ogen-go/ogen"
)

// Generate schemas:

//go:generate go run github.com/ogen-go/ogen/tools/mkformattest --output ../_testdata/positive/format_gen.json

// Generate JSON Schema:

//go:generate go run github.com/ogen-go/ogen/cmd/jschemagen --typename Spec --target ex_openapi/output.gen.go ../gen/_testdata/jsonschema/openapi30.json

// Fully supported:

//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_test_format --clean --generate-tests ../_testdata/positive/format_gen.json

//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_petstore --clean --generate-tests  ../_testdata/examples/petstore.yml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_firecracker --clean --generate-tests  ../_testdata/examples/firecracker.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_gotd --clean --generate-tests  ../_testdata/examples/gotd_bot_api.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_ent --clean --generate-tests  ../_testdata/examples/ent.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_route_params --clean --generate-tests  ../_testdata/positive/ex_route_params.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_manga --clean --generate-tests  ../_testdata/examples/manga.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_petstore_expanded --clean --generate-tests  ../_testdata/examples/petstore-expanded.yml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_telegram --clean --generate-tests  ../_testdata/examples/telegram_bot_api.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_2ch --clean --generate-tests  ../_testdata/examples/2ch.yml
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_tinkoff --clean --generate-tests  ../_testdata/examples/tinkoff.json

// Partially supported:

//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_k8s --clean --debug.noerr --generate-tests  ../_testdata/examples/k8s.json
//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --target ex_github --clean --infer-types --debug.noerr --generate-tests ../_testdata/examples/api.github.com.json
