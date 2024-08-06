package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/go-faster/yaml"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/location"
)

var errPanic = errors.New("panic")

func validateJSON(data []byte) error {
	d := jx.GetDecoder()
	d.ResetBytes(data)
	defer jx.PutDecoder(d)
	return d.Validate()
}

var gitPathRegex = regexp.MustCompile(`^/(?P<owner>[^/]+)/(?P<repo>[^/]+)/(-/blob|blob)/(?P<ref>[^/]+)/(?P<path>.*)$`)

func getRootURL(m FileMatch) (*url.URL, bool) {
	for _, external := range m.File.ExternalURLs {
		u, err := url.Parse(external.URL)
		if err != nil {
			continue
		}
		if !gitPathRegex.MatchString(u.Path) {
			continue
		}
		switch external.ServiceKind {
		case "GITHUB":
			u.Host = "raw.githubusercontent.com"
			u.Path = gitPathRegex.ReplaceAllString(u.Path, "/$owner/$repo/$ref/$path")
		case "GITLAB":
			u.Path = gitPathRegex.ReplaceAllString(u.Path, "/$owner/$repo/raw/$ref/$path")
		default:
			continue
		}
		return u, true
	}
	return nil, false
}

func worker(ctx context.Context, m FileMatch, r *Reporters, skipWrite bool) (rErr error) {
	data := []byte(m.File.Content)
	isYAML := strings.HasSuffix(m.File.Name, ".yml") || strings.HasSuffix(m.File.Name, ".yaml")
	hash := sha256.Sum256(data)

	defer func() {
		if rr := recover(); rr != nil {
			rErr = errPanic
			if err := r.report(ctx, Crash, Report{
				File:  m,
				Error: fmt.Sprintf("panic: %v", rr),
				Hash:  hash,
			}); err != nil {
				return
			}
		}
	}()

	f := file{
		data:   data,
		isYAML: isYAML,
		name:   m.File.Path,
		rootURL: &url.URL{
			Scheme: "jsonschema",
			Host:   "dummy",
		},
	}
	if u, ok := getRootURL(m); ok {
		f.rootURL = u
	}

	pse := generate(f, skipWrite)
	if pse != nil {
		if err := r.report(ctx, pse.stage, Report{
			File:           m,
			Error:          pse.Error(),
			NotImplemented: pse.notImpl,
			Hash:           hash,
		}); err != nil {
			return errors.Wrap(err, "report")
		}
		return errors.Wrap(pse, "generate")
	}

	if err := r.report(ctx, Good, Report{
		File: m,
		Hash: hash,
	}); err != nil {
		return errors.Wrap(err, "report")
	}
	return nil
}

type nopFs struct{}

func (n nopFs) WriteFile(string, []byte) error {
	return nil
}

func workerHTTPClient() *http.Client {
	dc := http.DefaultClient
	return &http.Client{
		Transport: filterTransport{
			next: http.DefaultTransport,
			allowed: map[string]struct{}{
				"raw.githubusercontent.com": {},
			},
		},
		CheckRedirect: dc.CheckRedirect,
		Jar:           dc.Jar,
		Timeout:       1 * time.Minute,
	}
}

type file struct {
	data    []byte
	isYAML  bool
	name    string
	rootURL *url.URL
}

func (f file) location() location.File {
	return location.NewFile(f.name, f.rootURL.String(), f.data)
}

func generate(f file, skipWrite bool) *GenerateError {
	if f.isYAML {
		var n *yaml.Node
		if err := yaml.Unmarshal(f.data, &n); err != nil {
			return &GenerateError{stage: InvalidYAML, err: err}
		}
	} else {
		if err := validateJSON(f.data); err != nil {
			return &GenerateError{stage: InvalidJSON, err: err}
		}
	}

	spec, err := ogen.Parse(f.data)
	if err != nil {
		return &GenerateError{stage: Unmarshal, err: err}
	}

	var (
		notImpl      []string
		firstNotImpl error
	)
	g, err := gen.NewGenerator(spec, gen.Options{
		Parser: gen.ParseOptions{
			InferSchemaType: true,
			AllowRemote:     true,
			RootURL:         f.rootURL,
			Remote: gen.RemoteOptions{
				HTTPClient: workerHTTPClient(),
				ReadFile: func(string) ([]byte, error) {
					return nil, errors.New("local file reference is not allowed")
				},
			},
			File: f.location(),
		},
		Generator: gen.GenerateOptions{
			IgnoreNotImplemented: []string{"all"},
			NotImplementedHook: func(name string, err error) {
				for _, existing := range notImpl {
					if existing == name {
						return
					}
				}
				if firstNotImpl == nil {
					firstNotImpl = err
				}
				notImpl = append(notImpl, name)
			},
		},
	})

	slices.Sort(notImpl)
	if err != nil {
		if _, ok := errors.Into[*gen.ErrParseSpec](err); ok {
			return &GenerateError{stage: Parse, notImpl: notImpl, err: err}
		}
		if _, ok := errors.Into[*gen.ErrBuildRouter](err); ok {
			return &GenerateError{stage: BuildRouter, notImpl: notImpl, err: err}
		}
		return &GenerateError{stage: BuildIR, notImpl: notImpl, err: err}
	}

	if !skipWrite {
		if err := g.WriteSource(nopFs{}, "api"); err != nil {
			if _, ok := errors.Into[*gen.ErrGoFormat](err); ok {
				return &GenerateError{stage: Format, notImpl: notImpl, err: err}
			}
			return &GenerateError{stage: Template, notImpl: notImpl, err: err}
		}
	}

	if len(notImpl) > 0 {
		return &GenerateError{stage: NotImplemented, notImpl: notImpl, err: firstNotImpl}
	}
	return nil
}
