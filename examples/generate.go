package examples

import (
	_ "github.com/go-faster/errors"

	_ "github.com/ogen-go/ogen"
)

// Fully supported:

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/petstore.yaml --target ex_petstore --clean
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/firecracker.json --target ex_firecracker --clean
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/gotd_bot_api.json --target ex_gotd --clean
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/tinkoff.json --target ex_tinkoff --clean
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/ent.json --target ex_ent --clean
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/ex_route_params.json --target ex_route_params --clean

// Partially supported:

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/manga.json --target ex_manga --clean --debug.noerr "unsupported content types"
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/petstore-expanded.yaml --target ex_petstore_expanded --clean --debug.noerr
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/k8s.json --target ex_k8s --clean --debug.noerr
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/api.github.com.json --target ex_github --clean --debug.noerr
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/telegram_bot_api.json --target ex_telegram --clean --debug.noerr
