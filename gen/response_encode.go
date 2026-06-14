package gen

import (
	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/gen/ir"
)

// sseServerResponseEncoding reports that SSE server response encoding is not implemented yet.
func sseServerResponseEncoding(op *ir.Operation) (any, error) {
	kind := "operation"
	if op.WebhookInfo != nil {
		kind = "webhook"
	}

	return nil, errors.Wrapf(
		&ErrNotImplemented{Name: "sse server response encoding"},
		"%s %q", kind, op.Name,
	)
}
