package ogen

import (
	"bytes"
	"encoding/json"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"gopkg.in/yaml.v3"
)

func wrapLineOffset(data []byte, err error) error {
	loc, ok := errors.Into[*json.UnmarshalTypeError](err)
	if !ok {
		return err
	}

	if loc.Offset < 0 || int64(len(data)) <= loc.Offset {
		return err
	}

	lines := data[:loc.Offset]
	// Lines count from 1.
	line := bytes.Count(lines, []byte("\n")) + 1
	lastNL := int64(bytes.LastIndexByte(lines, '\n'))
	column := loc.Offset - lastNL

	return errors.Wrapf(err, "line %d:%d", line, column)
}

// Parse parses JSON/YAML into OpenAPI Spec.
func Parse(data []byte) (s *Spec, err error) {
	s = &Spec{}
	if !jx.Valid(data) {
		if err := yaml.Unmarshal(data, s); err != nil {
			return nil, errors.Wrap(err, "yaml")
		}
	} else {
		if err := json.Unmarshal(data, s); err != nil {
			return nil, errors.Wrap(wrapLineOffset(data, err), "json")
		}
	}
	return s, nil
}
