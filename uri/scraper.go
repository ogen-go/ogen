package uri

import "fmt"

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

func (s *scraper) Value(v string) error {
	if s.typ == typeValue {
		return fmt.Errorf("multiple Value calls")
	}
	if s.typ != typeNotSet && s.typ != typeValue {
		return fmt.Errorf("encode value: already encoded as %s", s.typ)
	}

	s.typ = typeValue
	s.val = v
	return nil
}

func (s *scraper) Array(f func(Encoder) error) error {
	if s.typ != typeNotSet && s.typ != typeArray {
		return fmt.Errorf("encode array: already encoded as %s", s.typ)
	}

	s.typ = typeArray
	arr := &arrayScraper{}
	if err := f(arr); err != nil {
		return err
	}

	s.items = arr.items
	return nil
}

func (s *scraper) Field(field string, f func(Encoder) error) error {
	if s.typ != typeNotSet && s.typ != typeObject {
		return fmt.Errorf("encode object: already encoded as %s", s.typ)
	}

	s.typ = typeObject
	fenc := &fieldScraper{}
	if err := f(fenc); err != nil {
		return err
	}

	s.fields = append(s.fields, Field{
		Name:  field,
		Value: fenc.value,
	})
	return nil
}

type arrayScraper struct {
	set   bool
	items []string
}

func (e *arrayScraper) Value(v string) error {
	e.set = true
	e.items = append(e.items, v)
	return nil
}

func (e *arrayScraper) Array(_ func(Encoder) error) error {
	panic("nested arrays not allowed in path parameters")
}

func (e *arrayScraper) Field(_ string, _ func(Encoder) error) error {
	panic("nested objects not allowed in path parameters")
}

type fieldScraper struct {
	set   bool
	value string
}

func (e *fieldScraper) Value(v string) error {
	if e.set {
		panic("value already set")
	}
	e.value = v
	e.set = true
	return nil
}

func (e *fieldScraper) Array(_ func(Encoder) error) error {
	panic("nested arrays not allowed in path parameters")
}

func (e *fieldScraper) Field(_ string, _ func(Encoder) error) error {
	panic("nested objects not allowed in path parameters")
}
