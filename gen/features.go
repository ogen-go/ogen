package gen

import (
	"fmt"
	"slices"

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
	ClientRequestValidation = Feature{
		"client/request/validation",
		`Enables validation of client requests`,
	}
	ServerResponseValidation = Feature{
		"server/response/validation",
		`Enables validation of server responses`,
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
	ClientRequestValidation,
	ServerResponseValidation,
	OgenOtel,
	OgenUnimplemented,
	DebugExampleTests,
}
