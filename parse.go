package ogen

import (
	"encoding/json"
	"errors"

	"github.com/goccy/go-yaml"
	"github.com/ogen-go/jx"
)

func Parse(data []byte) (*Spec, error) {
	if !jx.Valid(data) {
		d, err := yaml.YAMLToJSON(data)
		if err != nil {
			return nil, err
		}
		data = d
	}
	if len(data) == 0 {
		return nil, errors.New("blank data")
	}

	s := &Spec{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}

	return s, nil
}
