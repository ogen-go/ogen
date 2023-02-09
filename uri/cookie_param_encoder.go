package uri

import (
	"net/http"
	"strings"
)

type cookieParamEncoder struct {
	*receiver
	paramName string
	explode   bool
	req       *http.Request
}

func (e *cookieParamEncoder) setCookie(val string) {
	e.req.AddCookie(&http.Cookie{
		Name:  e.paramName,
		Value: escapeCookie(val),
	})
}

func (e *cookieParamEncoder) serialize() {
	switch e.typ {
	case typeNotSet:
		return
	case typeValue:
		e.setCookie(e.val)
		return
	case typeArray:
		if e.explode {
			panic("cookie with explode: true not supported")
		}

		e.setCookie(strings.Join(e.items, ","))
		return
	case typeObject:
		if e.explode {
			panic("cookie with explode: true not supported")
		}

		const kvSep, fieldSep = ',', ','
		e.setCookie(encodeObject(kvSep, fieldSep, e.fields))
		return
	default:
		panic("unreachable")
	}
}
