package examples

import (
	_ "github.com/go-faster/errors"

	_ "github.com/ogen-go/ogen"
)

//go:generate go run github.com/ogen-go/ogen/cmd/ogen -v --clean --generate-tests --target ex_petstore          ../_testdata/examples/petstore.yml
