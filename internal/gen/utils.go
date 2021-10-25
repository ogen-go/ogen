package gen

import (
	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/internal/ast"
)

func forEachOps(item ogen.PathItem, f func(method string, op ogen.Operation) error) error {
	var err error
	handle := func(method string, op *ogen.Operation) {
		if err != nil || op == nil {
			return
		}
		err = f(method, *op)
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

func isUnderlyingPrimitive(s *ast.Schema) bool {
	switch s.Kind {
	case ast.KindPrimitive, ast.KindEnum:
		return true
	case ast.KindAlias:
		return isUnderlyingPrimitive(s.AliasTo)
	case ast.KindPointer:
		return isUnderlyingPrimitive(s.PointerTo)
	default:
		return false
	}
}
