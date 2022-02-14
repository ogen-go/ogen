package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

var (
	errPanic       = errors.New("panic")
	errInvalidJSON = errors.New("invalid json")
)

var bomPrefix = []byte{0xEF, 0xBB, 0xBF}

func worker(ctx context.Context, m FileMatch, r Reporters) (rErr error) {
	data := bytes.TrimPrefix([]byte(m.File.Content), bomPrefix)

	if !json.Valid(data) {
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
		case "write":
			ch = r.Write
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

func (p *GenerateError) Error() string {
	return p.err.Error()
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
		return &GenerateError{stage: "write", err: err}
	}
	return nil
}
