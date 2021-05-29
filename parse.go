package ogen

import (
	"encoding/json"
	"io"
)

func Parse(r io.Reader) (*Spec, error) {
	s := &Spec{}
	d := json.NewDecoder(r)
	if err := d.Decode(s); err != nil {
		return nil, err
	}

	return s, nil
}
