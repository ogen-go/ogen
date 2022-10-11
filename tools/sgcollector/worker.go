package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"golang.org/x/exp/slices"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
)

var (
	errPanic  = errors.New("panic")
	bomPrefix = []byte{0xEF, 0xBB, 0xBF}
)

func convertYAMLtoJSON(data []byte) (_ []byte, rErr error) {
	defer func() {
		if rr := recover(); rr != nil {
			rErr = errors.Errorf("panic: %#v", rr)
		}
	}()
	j, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func validateJSON(data []byte) error {
	d := jx.GetDecoder()
	d.ResetBytes(data)
	defer jx.PutDecoder(d)
	return d.Validate()
}

func worker(ctx context.Context, m FileMatch, r *Reporters) (rErr error) {
	data := bytes.TrimPrefix([]byte(m.File.Content), bomPrefix)
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

	pse := generate(data, isYAML)
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

func generate(data []byte, isYAML bool) *GenerateError {
	if isYAML {
		j, err := convertYAMLtoJSON(data)
		if err != nil {
			return &GenerateError{stage: InvalidYAML, err: err}
		}
		data = j
	}

	if err := validateJSON(data); err != nil {
		return &GenerateError{stage: InvalidJSON, err: err}
	}

	spec, err := ogen.Parse(data)
	if err != nil {
		return &GenerateError{stage: Unmarshal, err: err}
	}

	var (
		notImpl      []string
		firstNotImpl error
	)
	g, err := gen.NewGenerator(spec, gen.Options{
		InferSchemaType: true,
		AllowRemote:     true,
		Remote: gen.RemoteOptions{
			HTTPClient: workerHTTPClient(),
			ReadFile: func(string) ([]byte, error) {
				return nil, errors.New("local file reference is not allowed")
			},
		},
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

	if err := g.WriteSource(nopFs{}, "api"); err != nil {
		if _, ok := errors.Into[*gen.ErrGoFormat](err); ok {
			return &GenerateError{stage: Format, notImpl: notImpl, err: err}
		}
		return &GenerateError{stage: Template, notImpl: notImpl, err: err}
	}

	if len(notImpl) > 0 {
		return &GenerateError{stage: NotImplemented, notImpl: notImpl, err: firstNotImpl}
	}
	return nil
}
