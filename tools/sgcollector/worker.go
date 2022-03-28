package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
	"github.com/ogen-go/ogen/gen/genfs"
)

var errPanic = errors.New("panic")

var bomPrefix = []byte{0xEF, 0xBB, 0xBF}

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
	hash := sha256.Sum256(data[:])

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

	if err := generate(data, isYAML); err != nil {
		var pse *GenerateError
		if !errors.As(err, &pse) {
			return errors.Wrap(err, "invalid schema")
		}

		if err := r.report(ctx, pse.stage, Report{
			File:           m,
			Error:          err.Error(),
			NotImplemented: pse.notImpl,
			Hash:           hash,
		}); err != nil {
			return errors.Wrap(err, "report")
		}
		return errors.Wrap(err, "generate")
	}
	return nil
}

// GenerateError reports that generation failed.
type GenerateError struct {
	stage   Stage
	notImpl []string
	err     error
}

func (p *GenerateError) Unwrap() error {
	return p.err
}

func (p *GenerateError) Error() string {
	return fmt.Sprintf("%s: %s", p.stage, p.err)
}

func generate(data []byte, isYAML bool) error {
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

	var notImpl []string
	g, err := gen.NewGenerator(spec, gen.Options{
		InferSchemaType:      true,
		IgnoreNotImplemented: []string{"all"},
		NotImplementedHook: func(name string, err error) {
			for _, existing := range notImpl {
				if existing == name {
					return
				}
			}
			notImpl = append(notImpl, name)
		},
	})
	sort.Strings(notImpl)
	if err != nil {
		if as := new(gen.ErrParseSpec); errors.As(err, &as) {
			return &GenerateError{stage: Parse, notImpl: notImpl, err: err}
		}
		if as := new(gen.ErrBuildRouter); errors.As(err, &as) {
			return &GenerateError{stage: BuildRouter, notImpl: notImpl, err: err}
		}
		return &GenerateError{stage: BuildIR, notImpl: notImpl, err: err}
	}

	if err := g.WriteSource(genfs.CheckFS{}, "api"); err != nil {
		var pse *GenerateError
		if errors.As(err, &pse) {
			return err
		}
		return &GenerateError{stage: Template, notImpl: notImpl, err: err}
	}
	return nil
}
