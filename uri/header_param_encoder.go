package uri

import (
	"net/http"
	"strings"
)

type headerParamEncoder struct {
	*receiver
	paramName string
	explode   bool
	header    http.Header
}

func (e *headerParamEncoder) serialize() error {
	switch e.typ {
	case typeNotSet:
		return nil
	case typeValue:
		e.header.Set(e.paramName, e.val)
		return nil
	case typeArray:
		const sep = ","
		for _, val := range e.items {
			if err := checkNotContains(val, sep); err != nil {
				return err
			}
		}
		e.header.Set(e.paramName, strings.Join(e.items, sep))
		return nil
	case typeObject:
		var kvSep, fieldSep byte = ',', ','
		if e.explode {
			kvSep = '='
		}
		for _, f := range e.fields {
			if err := checkNotContains(f.Name, string(kvSep)); err != nil {
				return err
			}
			if err := checkNotContains(f.Value, string(fieldSep)); err != nil {
				return err
			}
		}
		e.header.Set(e.paramName, encodeObject(kvSep, fieldSep, e.fields))
		return nil
	default:
		panic("unreachable")
	}
}
