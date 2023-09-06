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
		if strings.EqualFold(e.paramName, "set-cookie") {
			// As per RFC6265:
			//
			// Origin servers SHOULD NOT fold multiple Set-Cookie header fields into
			// a single header field. The usual mechanism for folding HTTP headers
			// fields (i.e., as defined in RFC2616) might change the semantics of
			// the Set-Cookie header field because the %x2C (",") character is used
			// by Set-Cookie in a way that conflicts with such folding.
			for _, val := range e.items {
				e.header.Add(e.paramName, val)
			}
		} else {
			const sep = ","
			for _, val := range e.items {
				if err := checkNotContains(val, sep); err != nil {
					return err
				}
			}
			e.header.Set(e.paramName, strings.Join(e.items, sep))
		}
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
