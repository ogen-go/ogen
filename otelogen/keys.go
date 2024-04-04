package otelogen

import (
	"go.opentelemetry.io/otel/metric"
)

const (
	ClientRequestCount = "ogen.client.request_count" // Outgoing request count total
	ClientErrorsCount  = "ogen.client.errors_count"  // Outgoing errors total
	ClientDuration     = "ogen.client.duration"      // Outgoing end to end duration, milliseconds
)

func ClientRequestCountCounter(meter metric.Meter) (metric.Int64Counter, error) {
	return meter.Int64Counter(
		ClientRequestCount,
		metric.WithDescription("Outgoing request count total"),
		metric.WithUnit("{count}"),
	)
}

func ClientErrorsCountCounter(meter metric.Meter) (metric.Int64Counter, error) {
	return meter.Int64Counter(
		ClientErrorsCount,
		metric.WithDescription("Outgoing errors total"),
		metric.WithUnit("{count}"),
	)
}

func ClientDurationHistogram(meter metric.Meter) (metric.Float64Histogram, error) {
	return meter.Float64Histogram(
		ClientDuration,
		metric.WithDescription("Outgoing end to end duration"),
		metric.WithUnit("ms"),
	)
}

const (
	ServerRequestCount = "ogen.server.request_count" // Incoming request count total
	ServerErrorsCount  = "ogen.server.errors_count"  // Incoming errors total
	ServerDuration     = "ogen.server.duration"      // Incoming end to end duration, milliseconds
)

func ServerRequestCountCounter(meter metric.Meter) (metric.Int64Counter, error) {
	return meter.Int64Counter(
		ServerRequestCount,
		metric.WithDescription("Incoming request count total"),
		metric.WithUnit("{count}"),
	)
}

func ServerErrorsCountCounter(meter metric.Meter) (metric.Int64Counter, error) {
	return meter.Int64Counter(
		ServerErrorsCount,
		metric.WithDescription("Incoming errors total"),
		metric.WithUnit("{count}"),
	)
}

func ServerDurationHistogram(meter metric.Meter) (metric.Float64Histogram, error) {
	return meter.Float64Histogram(
		ServerDuration,
		metric.WithDescription("Incoming end to end duration"),
		metric.WithUnit("ms"),
	)
}
