package gen

// defaultImports returns a map of default imports for the generated code.
// The keys are the import paths, and the values are the aliases (empty string means no alias).
func defaultImports() map[string]string {
	return map[string]string{
		"bytes":           "",
		"context":         "",
		"encoding/base64": "",
		"fmt":             "",
		"io":              "",
		"math":            "",
		"math/big":        "",
		"math/bits":       "",
		"mime":            "",
		"mime/multipart":  "",
		"net":             "",
		"net/http":        "",
		"net/netip":       "",
		"net/url":         "",
		"regexp":          "",
		"sort":            "",
		"strconv":         "",
		"strings":         "",
		"sync":            "",
		"time":            "",

		"github.com/go-faster/errors":              "",
		"github.com/go-faster/jx":                  "",
		"github.com/google/uuid":                   "",
		"go.opentelemetry.io/otel":                 "",
		"go.opentelemetry.io/otel/attribute":       "",
		"go.opentelemetry.io/otel/codes":           "",
		"go.opentelemetry.io/otel/metric":          "",
		"go.opentelemetry.io/otel/semconv/v1.26.0": "semconv",
		"go.opentelemetry.io/otel/trace":           "",
		"go.uber.org/multierr":                     "",

		"github.com/ogen-go/ogen/conv":       "",
		"github.com/ogen-go/ogen/http":       "ht",
		"github.com/ogen-go/ogen/middleware": "",
		"github.com/ogen-go/ogen/json":       "",
		"github.com/ogen-go/ogen/ogenregex":  "",
		"github.com/ogen-go/ogen/ogenerrors": "",
		"github.com/ogen-go/ogen/otelogen":   "",
		"github.com/ogen-go/ogen/uri":        "",
		"github.com/ogen-go/ogen/validate":   "",
	}
}
