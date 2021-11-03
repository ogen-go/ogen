package uri

import (
	"github.com/ogen-go/errors"
)

type valueType string

func (v valueType) String() string { return string(v) }

const (
	typeNotSet valueType = "notSet"
	typeValue  valueType = "value"
	typeArray  valueType = "array"
	typeObject valueType = "object"
)

var _ Encoder = (*scraper)(nil)

// scraper is used to receive data from code generated types.
type scraper struct {
	typ    valueType
	val    string   // value type
	items  []string // array type
	fields []Field  // object type
}

func newScraper() *scraper {
	return &scraper{
		typ: typeNotSet,
	}
}

func (s *scraper) EncodeValue(v string) error {
	if s.typ == typeValue {
		return errors.New("multiple Value calls")
	}
	if s.typ != typeNotSet && s.typ != typeValue {
		return errors.Errorf("encode value: already encoded as %s", s.typ)
	}

	s.typ = typeValue
	s.val = v
	return nil
}

func (s *scraper) EncodeArray(f func(Encoder) error) error {
	if s.typ != typeNotSet && s.typ != typeArray {
		return errors.Errorf("encode array: already encoded as %s", s.typ)
	}

	s.typ = typeArray
	arr := &arrayScraper{}
	if err := f(arr); err != nil {
		return err
	}

	s.items = arr.items
	return nil
}

func (s *scraper) EncodeField(field string, f func(Encoder) error) error {
	if s.typ != typeNotSet && s.typ != typeObject {
		return errors.Errorf("encode object: already encoded as %s", s.typ)
	}

	s.typ = typeObject
	vs := &valueScraper{}
	if err := f(vs); err != nil {
		return err
	}

	if !vs.set {
		return nil
	}

	s.fields = append(s.fields, Field{
		Name:  field,
		Value: vs.value,
	})
	return nil
}

type arrayScraper struct {
	set   bool
	items []string
}

func (e *arrayScraper) EncodeValue(v string) error {
	e.set = true
	e.items = append(e.items, v)
	return nil
}

func (e *arrayScraper) EncodeArray(_ func(Encoder) error) error {
	panic("nested arrays not allowed in path parameters")
}

func (e *arrayScraper) EncodeField(_ string, _ func(Encoder) error) error {
	panic("nested objects not allowed in path parameters")
}

type valueScraper struct {
	set   bool
	value string
}

func (e *valueScraper) EncodeValue(v string) error {
	if e.set {
		panic("value already set")
	}
	e.value = v
	e.set = true
	return nil
}

func (e *valueScraper) EncodeArray(_ func(Encoder) error) error {
	panic("nested arrays not allowed")
}

func (e *valueScraper) EncodeField(_ string, _ func(Encoder) error) error {
	panic("nested objects not allowed")
}
