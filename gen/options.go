package gen

import (
	"context"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-faster/yaml"
	"go.uber.org/zap"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/urlpath"
	"github.com/ogen-go/ogen/jsonschema"
	"github.com/ogen-go/ogen/location"
	"github.com/ogen-go/ogen/openapi"
)

// Options is Generator options.
type Options struct {
	// Parser sets parser options.
	Parser ParseOptions `json:"parser" yaml:"parser"`

	// Generator sets generator options.
	Generator GenerateOptions `json:"generator" yaml:"generator"`

	// ExpandSpec is a path to expanded spec.
	ExpandSpec string `json:"expand" yaml:"expand"`

	// Logger to use.
	Logger *zap.Logger `json:"-" yaml:"-"`
}

// SetLocation sets File, RootURL and RemoteOptions using given path or URL
// and returns file data.
func (o *Options) SetLocation(p string, opts RemoteOptions) ([]byte, error) {
	return o.Parser.SetLocation(p, opts)
}

func (o *Options) setDefaults() {
	o.Parser.setDefaults()
	if o.Logger == nil {
		o.Logger = zap.NewNop()
	}
}

// RemoteOptions is remote reference resolver options.
type RemoteOptions = jsonschema.ExternalOptions

// ParseOptions sets parsing options.
type ParseOptions struct {
	// InferSchemaType enables type inference for schemas. Schema parser will try to detect schema type
	// by its properties.
	InferSchemaType bool `json:"infer_types" yaml:"infer_types"`
	// AllowRemote enables remote references resolving.
	//
	// See https://github.com/ogen-go/ogen/issues/385.
	AllowRemote bool `json:"allow_remote" yaml:"allow_remote"`
	// RootURL is root URL for remote references resolving.
	RootURL *url.URL `json:"-" yaml:"-"`
	// Remote is remote reference resolver options.
	Remote RemoteOptions `json:"-" yaml:"-"`
	// SchemaDepthLimit is maximum depth of schema generation. Default is 1000.
	SchemaDepthLimit int `json:"depth_limit" yaml:"depth_limit"`
	// File is the file that is being parsed.
	//
	// Used for error messages.
	File location.File `json:"-" yaml:"-"`
}

// SetLocation sets File, RootURL and RemoteOptions using given path or URL
// and returns file data.
func (o *ParseOptions) SetLocation(p string, opts RemoteOptions) ([]byte, error) {
	o.Remote = opts
	r := jsonschema.NewExternalResolver(opts)

	containsDrive := runtime.GOOS == "windows" && filepath.VolumeName(p) != ""
	if u, _ := url.Parse(p); u != nil && !containsDrive && u.Scheme != "" {
		switch u.Scheme {
		case "http", "https":
			_, fileName := path.Split(u.Path)

			// FIXME(tdakkota): pass context.
			data, err := r.Get(context.Background(), p)
			if err != nil {
				return nil, err
			}

			o.RootURL = u
			o.File = location.NewFile(fileName, p, data)
			// Guard against reading local files in remote mode.
			o.Remote.ReadFile = func(p string) ([]byte, error) {
				return nil, errors.New("local files are not supported in remote mode")
			}

			return data, nil
		case "file":
			toPath := opts.URLToFilePath
			if toPath == nil {
				toPath = urlpath.URLToFilePath
			}

			converted, err := toPath(u)
			if err != nil {
				return nil, errors.Wrap(err, "convert url to file path")
			}
			p = converted
		default:
			return nil, errors.Errorf("unsupported scheme %q", u.Scheme)
		}
	}
	p = filepath.Clean(p)

	abs, err := filepath.Abs(p)
	if err != nil {
		return nil, err
	}
	_, fileName := filepath.Split(p)

	readFile := o.Remote.ReadFile
	if readFile == nil {
		readFile = os.ReadFile
	}

	data, err := readFile(p)
	if err != nil {
		return nil, err
	}

	u, err := urlpath.URLFromFilePath(abs)
	if err != nil {
		return nil, errors.Wrap(err, "convert file path to url")
	}

	o.RootURL = u
	o.File = location.NewFile(fileName, p, data)
	return data, nil
}

func (o *ParseOptions) setDefaults() {
	if o.SchemaDepthLimit <= 0 {
		o.SchemaDepthLimit = defaultSchemaDepthLimit
	}
}

// GenerateOptions sets generator options.
type GenerateOptions struct {
	// Features sets generator features.
	Features *FeatureOptions `json:"features" yaml:"features"`

	// Filters contains filters to skip operations.
	Filters Filters `json:"filters" yaml:"filters"`

	// IgnoreNotImplemented contains ErrNotImplemented messages to ignore.
	IgnoreNotImplemented []string `json:"ignore_not_implemented" yaml:"ignore_not_implemented"`
	// NotImplementedHook is hook for ErrNotImplemented errors.
	NotImplementedHook func(name string, err error) `json:"-" yaml:"-"`

	// ConvenientErrors control Convenient Errors feature.
	//
	// Default value is `auto` (0), NewError handler will be generated if possible.
	//
	// If value > 0 forces feature. An error will be returned if generator is unable to find common error pattern.
	//
	// If value < 0 disables feature entirely.
	ConvenientErrors ConvenientErrors `json:"convenient_errors" yaml:"convenient_errors"`
	// ContentTypeAliases contains content type aliases.
	ContentTypeAliases ContentTypeAliases `json:"content_type_aliases" yaml:"content_type_aliases"`
}

// ConvenientErrors is an option type to control `Convenient Errors` feature.
type ConvenientErrors int

// IsDisabled whether Convenient Errors is disabled.
func (c ConvenientErrors) IsDisabled() bool {
	return c < 0
}

// IsForced whether Convenient Errors is forced.
func (c ConvenientErrors) IsForced() bool {
	return c > 0
}

// String implements fmt.Stringer.
func (c ConvenientErrors) String() string {
	switch {
	case c < 0:
		return "off"
	case c > 0:
		return "on"
	default:
		return "auto"
	}
}

// IsBoolFlag implements flag.boolFlag.
func (c *ConvenientErrors) IsBoolFlag() bool {
	return true
}

// UnmarshalYAML implements [yaml.Unmarshaler].
func (c *ConvenientErrors) UnmarshalYAML(n *yaml.Node) error {
	var value string
	if err := n.Decode(&value); err != nil {
		return err
	}
	return c.Set(value)
}

// Set implements flag.Value.
func (c *ConvenientErrors) Set(s string) error {
	switch s {
	case "auto":
		*c = 0
		return nil
	case "on", "true":
		*c = 1
		return nil
	case "off", "false":
		*c = -1
		return nil
	default:
		return errors.Errorf(`expected "on", "off" or "auto", got %q`, s)
	}
}

// ContentTypeAliases maps content type to concrete ogen encoding.
type ContentTypeAliases map[string]ir.Encoding

// String implements fmt.Stringer.
func (m ContentTypeAliases) String() string {
	var (
		b     strings.Builder
		first = true
	)
	for k, v := range m {
		if first {
			first = false
		} else {
			b.WriteString(",")
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v.String())
	}
	return b.String()
}

// Set implements flag.Value.
func (m *ContentTypeAliases) Set(value string) error {
	if *m == nil {
		*m = ContentTypeAliases{}
	}
	before, after, ok := strings.Cut(value, "=")
	if !ok {
		return errors.Errorf("invalid mapping %q", value)
	}
	(*m)[before] = ir.Encoding(after)
	return nil
}

// Filters contains filters to skip operations.
type Filters struct {
	PathRegex *regexp.Regexp
	Methods   []string
}

// UnmarshalYAML implements [yaml.Unmarshaler].
func (f *Filters) UnmarshalYAML(n *yaml.Node) error {
	var v struct {
		PathRegex string   `yaml:"path_regex"`
		Methods   []string `yaml:"methods"`
	}
	if err := n.Decode(&v); err != nil {
		return err
	}

	var err error
	f.PathRegex, err = regexp.Compile(v.PathRegex)
	if err != nil {
		return errors.Wrapf(err, "compile path regex %q", v.PathRegex)
	}
	f.Methods = v.Methods
	return nil
}

func (f Filters) accept(op *openapi.Operation) bool {
	if f.PathRegex != nil && !f.PathRegex.MatchString(op.Path.String()) {
		return false
	}

	if len(f.Methods) > 0 {
		return slices.ContainsFunc(f.Methods, func(m string) bool { return strings.EqualFold(m, op.HTTPMethod) })
	}

	return true
}
