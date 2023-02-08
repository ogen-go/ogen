package uri

import (
	"net/http"
)

type cookieParamEncoder struct {
	*receiver
	paramName string
	explode   bool
	req       *http.Request
}

func (e *cookieParamEncoder) serialize() {
	switch e.typ {
	case typeNotSet:
		return
	case typeValue:
		e.req.AddCookie(&http.Cookie{
			Name:  e.paramName,
			Value: escapeCookie(e.val),
		})
		return
	case typeArray:
		panic("cookie with array values is not implemented")
	case typeObject:
		panic("cookie with object values is not implemented")
	default:
		panic("unreachable")
	}
}
