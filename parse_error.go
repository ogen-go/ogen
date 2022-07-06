package ogen

import (
	"github.com/go-faster/errors"
	"github.com/go-json-experiment/json"

	ogenjson "github.com/ogen-go/ogen/json"
)

func unmarshal(data []byte, out any) error {
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
		return wrapLineOffset(begin, lines, err)
	}
	return nil
}

func wrapLineOffset(offset int64, lines ogenjson.Lines, err error) error {
	line, column, ok := lines.LineColumn(offset)
	if !ok {
		return err
	}
	return errors.Wrapf(err, "line %d:%d", line, column)
}
