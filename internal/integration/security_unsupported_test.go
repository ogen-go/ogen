package integration_test

import (
	"context"

	api "github.com/ogen-go/ogen/internal/integration/test_security_unsupported"
)

// supportedSecurity implements exactly the API that should remain after ogen
// discards incomplete security requirements and their unused schemes.
type supportedSecurity struct{}

var (
	_ api.Handler         = (*supportedSecurity)(nil)
	_ api.SecurityHandler = (*supportedSecurity)(nil)
	_ api.SecuritySource  = (*supportedSecurity)(nil)
)

func (*supportedSecurity) Preserved(context.Context) error {
	return nil
}

func (*supportedSecurity) HandleBearer(
	ctx context.Context,
	_ api.OperationName,
	_ api.Bearer,
) (context.Context, error) {
	return ctx, nil
}

func (*supportedSecurity) Bearer(context.Context, api.OperationName) (api.Bearer, error) {
	return api.Bearer{}, nil
}
