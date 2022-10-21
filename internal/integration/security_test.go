package integration

import (
	"context"

	api "github.com/ogen-go/ogen/internal/integration/sample_api"
)

type securityKey struct{}

func (s sampleAPIServer) HandleAPIKey(ctx context.Context, operationID string, t api.APIKey) (context.Context, error) {
	return context.WithValue(ctx, securityKey{}, t.APIKey), nil
}

func (s sampleAPIServer) APIKey(ctx context.Context, operationID string) (api.APIKey, error) {
	return api.APIKey{APIKey: "десять"}, nil
}

func (s sampleAPIServer) SecurityTest(ctx context.Context) (string, error) {
	return ctx.Value(securityKey{}).(string), nil
}
