package examples

import _ "golang.org/x/xerrors"

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/api.github.com.json --target ex_github --clean --debug.noerr
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/k8s.json --target ex_k8s --clean --debug.noerr
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/firecracker.json --target ex_firecracker --clean --debug.noerr
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/gotd_bot_api.json --target ex_gotd --clean
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/telegram_bot_api.json --target ex_telegram --clean --debug.noerr
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/nh.json --target ex_nh --clean --debug.noerr
