package examples

import _ "golang.org/x/xerrors"

//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/api.github.com.json --target ex_github --clean --debug.ignore-optionals --debug.ignore-unspecified
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/k8s.json --target ex_k8s --clean --debug.ignore-optionals --debug.ignore-unspecified
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --schema ../_testdata/firecracker.json --target ex_firecracker --clean --debug.ignore-optionals --debug.ignore-unspecified
