package main

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
)

var (
	errPanic       = errors.New("panic")
	errInvalidJSON = errors.New("invalid json")
)

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

func worker(ctx context.Context, m FileMatch, r Reporters) (rErr error) {
	data := bytes.TrimPrefix([]byte(m.File.Content), bomPrefix)

	if strings.HasSuffix(m.File.Name, ".yml") || strings.HasSuffix(m.File.Name, ".yaml") {
		j, err := convertYAMLtoJSON(data)
		if err != nil {
			if err := r.report(ctx, InvalidYAML, Report{
				File:  m,
				Error: err.Error(),
			}); err != nil {
				return errors.Wrap(err, "report")
			}
		}
		data = j
	}
	if !jx.Valid(data) {
		if err := r.report(ctx, InvalidJSON, Report{
			File:  m,
			Error: errInvalidJSON.Error(),
		}); err != nil {
			return errors.Wrap(err, "report")
		}
		return errInvalidJSON
	}

	defer func() {
		if rr := recover(); rr != nil {
			rErr = errPanic
			if err := r.report(ctx, Crash, Report{
				File:  m,
				Error: fmt.Sprintf("panic: %v", rr),
			}); err != nil {
				return
			}
		}
	}()
	err := generate(data)
	if err != nil {
		var pse *GenerateError
		if !errors.As(err, &pse) {
			return errors.Wrap(err, "invalid schema")
		}

		if err := r.report(ctx, pse.stage, Report{
			File:  m,
			Error: err.Error(),
		}); err != nil {
			return errors.Wrap(err, "report")
		}
		return errors.Wrap(err, "generate")
	}
	return nil
}

// GenerateError reports that generation failed.
type GenerateError struct {
	stage Stage
	err   error
}

func (p *GenerateError) Unwrap() error {
	return p.err
}

func (p *GenerateError) Error() string {
	return p.err.Error()
}

type fmtFs struct{}

func (n fmtFs) WriteFile(baseName string, source []byte) error {
	_, err := format.Source(source)
	if err != nil {
		return &GenerateError{stage: Format, err: err}
	}
	return nil
}

func generate(data []byte) error {
	spec, err := ogen.Parse(data)
	if err != nil {
		return &GenerateError{stage: Parse, err: err}
	}

	g, err := gen.NewGenerator(spec, gen.Options{
		InferSchemaType:      true,
		IgnoreNotImplemented: []string{"all"},
	})
	if err != nil {
		return &GenerateError{stage: Build, err: err}
	}

	if err := g.WriteSource(fmtFs{}, "api"); err != nil {
		var pse *GenerateError
		if errors.As(err, &pse) {
			return err
		}
		return &GenerateError{stage: Template, err: err}
	}
	return nil
}
