package otelogen

const (
	ClientRequestCount = "ogen.client.request_count" // Outgoing request count total
	ClientErrorsCount  = "ogen.client.errors_count"  // Outgoing errors total
	ClientDuration     = "ogen.client.duration"      // Outgoing end to end duration, microseconds
)

const (
	ServerRequestCount = "ogen.server.request_count" // Incoming request count total
	ServerErrorsCount  = "ogen.server.errors_count"  // Incoming errors total
	ServerDuration     = "ogen.server.duration"      // Incoming end to end duration, microseconds
)
