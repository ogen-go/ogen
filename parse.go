package ogen

import (
	"encoding/json"
	"io"

	"github.com/goccy/go-yaml"
	"github.com/ogen-go/jx"
)

func Parse(r io.Reader) (*Spec, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if !jx.Valid(data) {
		d, err := yaml.YAMLToJSON(data)
		if err != nil {
			return nil, err
		}
		data = d
	}

	s := &Spec{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}

	return s, nil
}
