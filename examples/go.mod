module github.com/ogen-go/ogen/examples

go 1.17

require (
	github.com/go-chi/chi/v5 v5.0.5
	github.com/google/uuid v1.3.0
	github.com/ogen-go/ogen v0.0.0
	go.opentelemetry.io/otel v1.1.0
	go.opentelemetry.io/otel/metric v0.24.0
	go.opentelemetry.io/otel/trace v1.1.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

require (
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/ogen-go/json v0.0.0-20211028115912-1018eae39492 // indirect
)

replace github.com/ogen-go/ogen v0.0.0 => ./..
