package otelogen

import (
	"go.opentelemetry.io/otel/attribute"
)

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
