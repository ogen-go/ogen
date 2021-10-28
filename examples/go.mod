module github.com/ogen-go/ogen/examples

go 1.17

require (
	github.com/go-chi/chi/v5 v5.0.5
	github.com/google/uuid v1.3.0
	github.com/ogen-go/ogen v0.0.0
	go.opentelemetry.io/otel v1.1.0
	go.opentelemetry.io/otel/trace v1.1.0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

require (
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
)

replace github.com/ogen-go/ogen v0.0.0 => ./..
