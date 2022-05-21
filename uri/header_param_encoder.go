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

func (e *headerParamEncoder) serialize() {
	switch e.typ {
	case typeNotSet:
		return
	case typeValue:
		e.header.Set(e.paramName, e.val)
		return
	case typeArray:
		e.header.Set(e.paramName, strings.Join(e.items, ","))
		return
	case typeObject:
		kvSep, fieldSep := ',', ','
		if e.explode {
			kvSep = '='
		}
		e.header.Set(e.paramName, encodeObject(kvSep, fieldSep, e.fields))
		return
	default:
		panic("unreachable")
	}
}
