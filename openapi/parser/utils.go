package parser

import (
	"github.com/go-faster/errors"
	"github.com/ogen-go/ogen"
)

func forEachOps(item *ogen.PathItem, f func(method string, op ogen.Operation) error) error {
	var err error
	handle := func(method string, op *ogen.Operation) {
		if err != nil || op == nil {
			return
		}

		err = f(method, *op)
		if err != nil {
			err = errors.Wrap(err, method)
		}
	}

	handle("get", item.Get)
	handle("put", item.Put)
	handle("post", item.Post)
	handle("delete", item.Delete)
	handle("options", item.Options)
	handle("head", item.Head)
	handle("patch", item.Patch)
	handle("trace", item.Trace)
	return err
}
