// Package ogen implements OpenAPI v3 code generation.
package ogen

import (
	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

// Parse parses JSON/YAML into OpenAPI Spec.
func Parse(data []byte) (s *Spec, err error) {
	s = &Spec{}
	if !jx.Valid(data) {
		if err := unmarshalYAML(data, s); err != nil {
			return nil, errors.Wrap(err, "yaml")
		}
	} else {
		if err := unmarshalJSON(data, s); err != nil {
			return nil, errors.Wrap(err, "json")
		}
	}
	s.Init()
	return s, nil
}
