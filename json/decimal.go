package json

import (
	"github.com/go-faster/errors"
	"github.com/go-faster/jx"
	"github.com/shopspring/decimal"
)

// EncodeDecimal encodes decimal.Decimal to json.
func EncodeDecimal(e *jx.Encoder, v decimal.Decimal) {
	e.Num(jx.Num(v.String()))
}

// DecodeDecimal decodes decimal.Decimal from json.
func DecodeDecimal(d *jx.Decoder) (decimal.Decimal, error) {
	n, err := d.Num()
	if err != nil {
		return decimal.Decimal{}, err
	}
	v, err := decimal.NewFromString(n.String())
	if err != nil {
		return decimal.Decimal{}, errors.Wrap(err, "invalid decimal")
	}
	return v, nil
}

// EncodeStringDecimal encodes decimal.Decimal to json string.
func EncodeStringDecimal(e *jx.Encoder, v decimal.Decimal) {
	e.Str(v.String())
}

// DecodeStringDecimal decodes decimal.Decimal from json string.
func DecodeStringDecimal(d *jx.Decoder) (decimal.Decimal, error) {
	s, err := d.Str()
	if err != nil {
		return decimal.Decimal{}, err
	}
	v, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Decimal{}, errors.Wrap(err, "invalid decimal string")
	}
	return v, nil
}
