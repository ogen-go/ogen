package examples

import (
	_ "golang.org/x/xerrors"

	_ "github.com/ogen-go/ogen"
)

// Almost fully supported, except optional params:

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/petstore.yaml --target ex_petstore --clean --debug.noerr "optional params"
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/firecracker.json --target ex_firecracker --clean --debug.noerr "optional params"
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/gotd_bot_api.json --target ex_gotd --clean --debug.noerr "optional params"
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/manga.json --target ex_manga --clean --debug.noerr "optional params"

// Partially supported:

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/petstore-expanded.yaml --target ex_petstore_expanded --clean --debug.noerr
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/k8s.json --target ex_k8s --clean --debug.noerr
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/api.github.com.json --target ex_github --clean --debug.noerr
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/telegram_bot_api.json --target ex_telegram --clean --debug.noerr
