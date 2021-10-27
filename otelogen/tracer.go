package otelogen

import (
	"go.opentelemetry.io/otel/attribute"
)

type Tracer struct {
	cfg config
}

func New(opts ...Option) *Tracer {
	return &Tracer{
		cfg: newConfig(opts...),
	}
}

const (
	// OperationIDKey by OpenAPI specification.
	OperationIDKey = attribute.Key("oas.operation")
)

// OperationID attribute.
func OperationID(v string) attribute.KeyValue {
	return attribute.KeyValue{
		Key:   OperationIDKey,
		Value: attribute.StringValue(v),
	}
}
