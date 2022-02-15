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
			select {
			case <-ctx.Done():
				return
			case r.InvalidYAML <- Report{
				File:  m,
				Error: err.Error(),
			}:
				return errors.Wrap(err, "convert to JSON")
			}
		}
		data = j
	}
	if !jx.Valid(data) {
		select {
		case <-ctx.Done():
			return
		case r.InvalidJSON <- Report{
			File:  m,
			Error: errInvalidJSON.Error(),
		}:
		}
		return errInvalidJSON
	}

	defer func() {
		if rr := recover(); rr != nil {
			rErr = errPanic
			select {
			case <-ctx.Done():
				return
			case r.Crash <- Report{
				File:  m,
				Error: fmt.Sprintf("panic: %v", rr),
			}:
			}
		}
	}()
	err := generate(data)
	if err != nil {
		var pse *GenerateError
		if !errors.As(err, &pse) {
			return errors.Wrap(err, "invalid schema")
		}

		ch := r.Parse
		switch pse.stage {
		case "build":
			ch = r.Build
		case "template":
			ch = r.Template
		case "format":
			ch = r.Format
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- Report{
			File:  m,
			Error: err.Error(),
		}:
			return errors.Wrap(err, "generate")
		}
	}
	return nil
}

// GenerateError reports that generation failed.
type GenerateError struct {
	stage string
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
		return &GenerateError{stage: "format", err: err}
	}
	return nil
}

func generate(data []byte) error {
	spec, err := ogen.Parse(data)
	if err != nil {
		return &GenerateError{stage: "parse", err: err}
	}

	g, err := gen.NewGenerator(spec, gen.Options{
		InferSchemaType:      true,
		IgnoreNotImplemented: []string{"all"},
	})
	if err != nil {
		return &GenerateError{stage: "build", err: err}
	}

	if err := g.WriteSource(fmtFs{}, "api"); err != nil {
		var pse *GenerateError
		if errors.As(err, &pse) {
			return err
		}
		return &GenerateError{stage: "template", err: err}
	}
	return nil
}
