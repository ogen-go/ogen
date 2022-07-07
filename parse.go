package ogen

import (
	"github.com/go-faster/errors"
	"github.com/go-json-experiment/json"
	"gopkg.in/yaml.v3"

	ogenjson "github.com/ogen-go/ogen/json"
)

func unmarshalYAML(data []byte, out any) error {
	return yaml.Unmarshal(data, out)
}

func unmarshalJSON(data []byte, out any) error {
	var lines ogenjson.Lines
	lines.Collect(data)

	begin := int64(-1)
	opts := json.UnmarshalOptions{
		Unmarshalers: json.NewUnmarshalers(
			json.UnmarshalFuncV2(func(opts json.UnmarshalOptions, d *json.Decoder, t any) (rerr error) {
				begin = d.InputOffset()
				return json.SkipFunc
			}),
			ogenjson.LocationUnmarshaler(lines),
		),
	}

	if err := opts.Unmarshal(json.DecodeOptions{}, data, out); err != nil {
		line, column, ok := lines.LineColumn(begin)
		if !ok {
			return err
		}
		return errors.Wrapf(err, "line %d:%d", line, column)
	}
	return nil
}
