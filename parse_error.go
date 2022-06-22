package ogen

import (
	"bytes"
	"encoding/json"

	"github.com/go-faster/errors"
)

func wrapLineOffset(data []byte, err error) error {
	loc, ok := errors.Into[*json.UnmarshalTypeError](err)
	if !ok {
		return err
	}

	if loc.Offset < 0 || int64(len(data)) <= loc.Offset {
		return err
	}

	lines := data[:loc.Offset]
	// Lines count from 1.
	line := bytes.Count(lines, []byte("\n")) + 1
	lastNL := int64(bytes.LastIndexByte(lines, '\n'))
	column := loc.Offset - lastNL

	return errors.Wrapf(err, "line %d:%d", line, column)
}
