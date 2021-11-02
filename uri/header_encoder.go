package uri

import "strings"

type HeaderEncoder struct {
	*scraper
	explode bool
}

type HeaderEncoderConfig struct {
	Explode bool
}

func NewHeaderEncoder(cfg HeaderEncoderConfig) *HeaderEncoder {
	return &HeaderEncoder{
		scraper: newScraper(),
		explode: cfg.Explode,
	}
}

func (e *HeaderEncoder) Result() (string, bool) {
	switch e.typ {
	case typeNotSet:
		return "", false
	case typeValue:
		return e.val, true
	case typeArray:
		return strings.Join(e.items, ","), true
	case typeObject:
		kvSep, fieldSep := ',', ','
		if e.explode {
			kvSep = '='
		}
		return encodeObject(kvSep, fieldSep, e.fields), true
	default:
		panic("unreachable")
	}
}
