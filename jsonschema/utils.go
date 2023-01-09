package jsonschema

import (
	"encoding/json"

	"github.com/go-faster/jx"
)

func getRawSchemaFields(s *RawSchema) ([]string, error) {
	bs, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var fields []string

	d := jx.DecodeBytes(bs)
	if err := d.Obj(func(d *jx.Decoder, key string) error {
		fields = append(fields, key)
		return d.Skip()
	}); err != nil {
		return nil, err
	}

	return fields, nil
}
