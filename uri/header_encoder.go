package uri

import "strings"

type HeaderEncoder struct {
	explode bool
}

type HeaderEncoderConfig struct {
	Explode bool
}

func (e HeaderEncoder) EncodeString(v string) string {
	return v
}

func (e HeaderEncoder) EncodeStrings(v []string) string {
	return strings.Join(v, ",")
}

func (e HeaderEncoder) EncodeObject(fields []Field) string {
	kvSep, fieldSep := ',', ','
	if e.explode {
		kvSep = '='
	}
	return encodeObject(kvSep, fieldSep, fields)
}
