package gen

import (
	"errors"
)

// errSkipSchema allows to skip generation.
var errSkipSchema = errors.New("skip")
