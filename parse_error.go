package ogen

import (
	"bytes"

	"github.com/go-faster/errors"
	"github.com/go-json-experiment/json"
)

func unmarshal(data []byte, out any) error {
	begin := int64(-1)
	opts := json.UnmarshalOptions{
		Unmarshalers: json.NewUnmarshalers(
			json.UnmarshalFuncV2(func(opts json.UnmarshalOptions, d *json.Decoder, t any) (rerr error) {
				begin = d.InputOffset()
				return json.SkipFunc
			}),
		),
	}

	if err := opts.Unmarshal(json.DecodeOptions{}, data, out); err != nil {
		return wrapLineOffset(begin, data, err)
	}
	return nil
}

func wrapLineOffset(offset int64, data []byte, err error) error {
	if offset < 0 || int64(len(data)) <= offset {
		return err
	}

	{
		unread := data[offset:]
		trimmed := bytes.TrimLeft(unread, "\x20\t\r\n,:")
		if len(trimmed) != len(unread) {
			// Skip leading whitespace, because decoder does not do it.
			offset += int64(len(unread) - len(trimmed))
		}
	}

	lines := data[:offset]
	// Lines count from 1.
	line := bytes.Count(lines, []byte("\n")) + 1
	lastNL := int64(bytes.LastIndexByte(lines, '\n'))
	column := offset - lastNL

	return errors.Wrapf(err, "line %d:%d", line, column)
}
