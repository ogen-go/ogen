module github.com/ogen-go/ogen/examples

go 1.18

require (
	github.com/go-faster/errors v0.6.1
	github.com/go-faster/jx v0.38.1
	github.com/google/uuid v1.3.0
	github.com/ogen-go/ogen v0.0.0
	github.com/stretchr/testify v1.8.0
	go.opentelemetry.io/otel v1.8.0
	go.opentelemetry.io/otel/metric v0.31.0
	go.opentelemetry.io/otel/trace v1.8.0
	go.uber.org/multierr v1.8.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-faster/yamlx v0.2.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f // indirect
	golang.org/x/sys v0.0.0-20220622161953-175b2fd9d664 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/ogen-go/ogen v0.0.0 => ./..
	gopkg.in/yaml.v3 v3.0.1 => github.com/go-faster/yamlx v0.0.0-20220711115722-810b8bfdedac
)
