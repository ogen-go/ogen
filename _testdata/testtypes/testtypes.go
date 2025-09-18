package testtypes

import (
	"strconv"

	"github.com/go-faster/jx"
)

type StringOgen struct{ Value string }

func (o *StringOgen) Encode(e *jx.Encoder) {
	e.Str(o.Value)
}

func (o *StringOgen) Decode(d *jx.Decoder) error {
	s, err := d.Str()
	if err != nil {
		return err
	}
	o.Value = s
	return nil
}

type NumberOgen struct{ Value float64 }

func (o *NumberOgen) Encode(e *jx.Encoder) {
	e.Float64(o.Value)
}

func (o *NumberOgen) Decode(d *jx.Decoder) error {
	s, err := d.Float64()
	if err != nil {
		return err
	}
	o.Value = s
	return nil
}

type StringJSON struct{ Value string }

func (j *StringJSON) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(j.Value)), nil
}

func (j *StringJSON) UnmarshalJSON(data []byte) error {
	s, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	j.Value = s
	return nil
}

type NumberJSON struct{ Value float64 }

func (j *NumberJSON) MarshalJSON() ([]byte, error) {
	return strconv.AppendFloat(nil, j.Value, 'f', -1, 64), nil
}

func (j *NumberJSON) UnmarshalJSON(data []byte) error {
	n, err := strconv.ParseFloat(string(data), 64)
	j.Value = n
	return err
}

type Text struct{ Value string }

func (t *Text) MarshalText() ([]byte, error) {
	return []byte(t.Value), nil
}

func (t *Text) UnmarshalText(data []byte) error {
	t.Value = string(data)
	return nil
}

type Binary struct{ Value string }

func (b *Binary) MarshalBinary() ([]byte, error) {
	return []byte(b.Value), nil
}

func (b *Binary) UnmarshalBinary(data []byte) error {
	b.Value = string(data)
	return nil
}

type String string

type Number float64
