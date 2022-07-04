// Package ogen implements OpenAPI v3 code generation.
package ogen

import (
	"github.com/ghodss/yaml"
	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
)

// Parse parses JSON/YAML into OpenAPI Spec.
func Parse(data []byte) (s *Spec, err error) {
	s = &Spec{}
	if !jx.Valid(data) {
		d, err := yaml.YAMLToJSON(data)
		if err != nil {
			return nil, errors.Wrap(err, "yaml")
		}
		data = d
	}
	if err := unmarshal(data, s); err != nil {
		return nil, errors.Wrap(err, "json")
	}
	return s, nil
}
