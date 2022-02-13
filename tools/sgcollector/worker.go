package main

import (
	"context"
	"encoding/json"
	"go/format"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen"
	"github.com/ogen-go/ogen/gen"
)

type fmtFs struct{}

func (n fmtFs) WriteFile(baseName string, source []byte) error {
	_, err := format.Source(source)
	return err
}

var errPanic = errors.New("panic")

func worker(ctx context.Context, m FileMatch, invalidSchema chan<- Report, crashSchema chan<- Report) (rErr error) {
	data := []byte(m.File.Content)

	if !json.Valid(data) {
		return errors.New("invalid json")
	}

	defer func() {
		if rr := recover(); rr != nil {
			rErr = errPanic
			select {
			case <-ctx.Done():
				return
			case crashSchema <- Report{
				File:  m,
				Error: rErr.Error(),
				Data:  data,
			}:
			}
		}
	}()
	err := generate(data)
	if err != nil {
		var pse *ParseSpecError
		if errors.As(err, &pse) {
			return errors.Wrap(err, "invalid schema")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case invalidSchema <- Report{
			File:  m,
			Error: err.Error(),
			Data:  data,
		}:
			return errors.Wrap(err, "parse spec")
		}
	}
	return nil
}

// ParseSpecError reports that specification parsing failed.
type ParseSpecError struct {
	err error
}

func (p *ParseSpecError) Error() string {
	return p.err.Error()
}

func generate(data []byte) error {
	spec, err := ogen.Parse(data)
	if err != nil {
		return &ParseSpecError{err: err}
	}

	g, err := gen.NewGenerator(spec, gen.Options{
		InferSchemaType:      true,
		IgnoreNotImplemented: []string{"all"},
	})
	if err != nil {
		return errors.Wrap(err, "build IR")
	}

	if err := g.WriteSource(fmtFs{}, "api"); err != nil {
		return errors.Wrap(err, "write source")
	}
	return nil
}
