package ogen

import (
	"github.com/go-faster/errors"
	"github.com/go-json-experiment/json"

	ogenjson "github.com/ogen-go/ogen/json"
)

func unmarshal(data []byte, out any) error {
	begin := int64(-1)
	opts := json.UnmarshalOptions{
		Unmarshalers: json.NewUnmarshalers(
			json.UnmarshalFuncV2(func(opts json.UnmarshalOptions, d *json.Decoder, t any) (rerr error) {
				begin = d.InputOffset()
				return json.SkipFunc
			}),
			ogenjson.LocationUnmarshaler(),
		),
	}

	if err := opts.Unmarshal(json.DecodeOptions{}, data, out); err != nil {
		return wrapLineOffset(begin, data, err)
	}
	return nil
}

func wrapLineOffset(offset int64, data []byte, err error) error {
	line, column, ok := ogenjson.LineColumn(offset, data)
	if !ok {
		return err
	}
	return errors.Wrapf(err, "line %d:%d", line, column)
}
