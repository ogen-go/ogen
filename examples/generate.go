package examples

import (
	_ "github.com/go-faster/errors"

	_ "github.com/ogen-go/ogen"
)

// Generate schemas:

//go:generate go run github.com/ogen-go/ogen/tools/mkformattest --output ../_testdata/positive/test_format.json

// Fully supported:

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/test_format.json --target ex_test_format --clean --generate-tests

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/petstore.yaml --target ex_petstore --clean --generate-tests
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/firecracker.json --target ex_firecracker --clean --generate-tests
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/gotd_bot_api.json --target ex_gotd --clean --generate-tests
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/tinkoff.json --target ex_tinkoff --clean --generate-tests
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/ent.json --target ex_ent --clean --generate-tests
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/ex_route_params.json --target ex_route_params --clean --generate-tests

// Partially supported:

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/manga.json --target ex_manga --clean --debug.noerr "unsupported content types" --generate-tests
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/petstore-expanded.yaml --target ex_petstore_expanded --clean --debug.noerr --generate-tests
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/k8s.json --target ex_k8s --clean --debug.noerr --generate-tests
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/api.github.com.json --target ex_github --clean --infer-types --debug.noerr --generate-tests
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/positive/telegram_bot_api.json --target ex_telegram --clean --debug.noerr --generate-tests
