package gen

import (
	"fmt"
	"go/version"
	"slices"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"
)

// Feature is an ogen feature.
type Feature struct {
	Name        string
	Description string
}

// FeatureOptions is features Options.
type FeatureOptions struct {
	Enable     FeatureSet `json:"enable" yaml:"enable"`
	Disable    FeatureSet `json:"disable" yaml:"disable"`
	DisableAll bool       `json:"disable_all" yaml:"disable_all"`
}

// Build returns final set.
func (cfg *FeatureOptions) Build() (set FeatureSet, _ error) {
	if cfg == nil {
		cfg = &FeatureOptions{}
	}

	set = make(FeatureSet)
	if !cfg.DisableAll {
		for _, f := range DefaultFeatures {
			if err := set.Enable(f.Name); err != nil {
				panic(fmt.Sprintf("bad default feature %q", f.Name))
			}
		}
	}
	for name := range cfg.Disable {
		set.Disable(name)
	}
	for name := range cfg.Enable {
		if err := set.Enable(name); err != nil {
			return set, err
		}
	}
	return set, nil
}

// FeatureSet is set of [Feature] names.
type FeatureSet map[string]struct{}

// Enable adds a feature to set.
func (s *FeatureSet) Enable(name string) error {
	if *s == nil {
		*s = make(FeatureSet)
	}
	if !slices.ContainsFunc(
		AllFeatures,
		func(f Feature) bool { return f.Name == name },
	) {
		return errors.Errorf("unknown feature %q", name)
	}
	(*s)[name] = struct{}{}
	return nil
}

// minGoVersion maps features to the minimal Go version they require. Such
// features are enabled automatically if the target Go version is new enough,
// unless they are explicitly disabled.
var minGoVersion = map[Feature]string{
	FormatStdlibUUID: "go1.27",
}

// enableForGoVersion enables features available for the target Go version,
// like "1.27" or "go1.27.1". Empty version enables nothing.
func (s *FeatureSet) enableForGoVersion(cfg *FeatureOptions, goVersion string) error {
	if goVersion == "" {
		return nil
	}
	v := goVersion
	if !strings.HasPrefix(v, "go") {
		v = "go" + v
	}
	if !version.IsValid(v) {
		return errors.Errorf("invalid Go version %q", goVersion)
	}
	for f, minVersion := range minGoVersion {
		if version.Compare(v, minVersion) < 0 || cfg.disabled(f) {
			continue
		}
		if err := s.Enable(f.Name); err != nil {
			return err
		}
	}
	return nil
}

// disabled reports whether feature is explicitly disabled.
func (cfg *FeatureOptions) disabled(f Feature) bool {
	if cfg == nil {
		return false
	}
	if cfg.DisableAll {
		return true
	}
	_, ok := cfg.Disable[f.Name]
	return ok
}

// Disable removes a feature from set.
func (s *FeatureSet) Disable(name string) {
	delete(*s, name)
}

// Has whether if set has given feature.
func (s FeatureSet) Has(feature Feature) bool {
	_, ok := s[feature.Name]
	return ok
}

// UnmarshalYAML implements [yaml.Unmarshaler].
func (s *FeatureSet) UnmarshalYAML(n *yaml.Node) error {
	var value []string
	if err := n.Decode(&value); err != nil {
		return err
	}

	*s = make(FeatureSet, len(value))
	for _, name := range value {
		if err := s.Enable(name); err != nil {
			return err
		}
	}

	return nil
}

var (
	PathsClient = Feature{
		"paths/client",
		`Enables paths client generation`,
	}
	PathsServer = Feature{
		"paths/server",
		`Enables paths server generation`,
	}
	WebhooksClient = Feature{
		"webhooks/client",
		`Enables webhooks client generation`,
	}
	WebhooksServer = Feature{
		"webhooks/server",
		`Enables webhooks server generation`,
	}
	ClientSecurityReentrant = Feature{
		"client/security/reentrant",
		`Enables client usage in security source implementations`,
	}
	ClientRequestOptions = Feature{
		"client/request/options",
		`Enables function options for client requests`,
	}
	ClientRequestValidation = Feature{
		"client/request/validation",
		`Enables validation of client requests`,
	}
	ClientEditors = Feature{
		"client/editors",
		`Enables editors function options for client`,
	}
	ServerResponseValidation = Feature{
		"server/response/validation",
		`Enables validation of server responses`,
	}
	ServerSpec = Feature{
		"server/spec",
		`Embeds the OpenAPI document into generated code and generates SpecJSON, SpecYAML, ` +
			`SpecHandler and the WithDocs server option`,
	}
	ServerSwaggerUI = Feature{
		"server/swagger_ui",
		`Generates SwaggerUIHandler and serves Swagger UI via WithDocs. Implies server/spec`,
	}
	OgenOtel = Feature{
		"ogen/otel",
		`Enables OpenTelemetry integration`,
	}
	OgenUnimplemented = Feature{
		"ogen/unimplemented",
		`Enables stub Handler generation`,
	}
	DebugExampleTests = Feature{
		"debug/example_tests",
		`Enables example tests generation`,
	}
	SchemaEmbedStructAliases = Feature{
		"schema/embed_struct_aliases",
		`Generates aliases of object schemas as structs embedding the original type ` +
			`(type Foo struct{ Bar }) instead of defined types (type Foo Bar), ` +
			`so methods of the original type are promoted. Opt-in, since it changes generated API.`,
	}
	FormatStdlibUUID = Feature{
		"format/stdlib_uuid",
		`Uses the standard library uuid.UUID (Go 1.27+) for "format: uuid" instead of ` +
			`github.com/google/uuid. Enabled automatically if the target Go version is 1.27 or newer, ` +
			`disable it to keep github.com/google/uuid.`,
	}
	NamingCamelInitialisms = Feature{
		"naming/camel_initialisms",
		`Applies initialism rules (ID, URL, HTTP, ...) to camelCase identifiers, ` +
			`e.g. "userId" becomes "UserID". Opt-in, since it changes generated names.`,
	}
)

// DefaultFeatures defines default ogen features.
var DefaultFeatures = []Feature{
	PathsClient,
	PathsServer,
	WebhooksClient,
	WebhooksServer,
	OgenOtel,
	OgenUnimplemented,
}

// AllFeatures contains all ogen features.
var AllFeatures = []Feature{
	PathsClient,
	PathsServer,
	WebhooksClient,
	WebhooksServer,
	ClientSecurityReentrant,
	ClientRequestOptions,
	ClientRequestValidation,
	ClientEditors,
	ServerResponseValidation,
	ServerSpec,
	ServerSwaggerUI,
	OgenOtel,
	OgenUnimplemented,
	DebugExampleTests,
	SchemaEmbedStructAliases,
	FormatStdlibUUID,
	NamingCamelInitialisms,
}
