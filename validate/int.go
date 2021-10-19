package validate

import (
	"fmt"
)

// Int validates integers.
type Int struct {
	MultipleOf    int
	MultipleOfSet bool

	Min          int64
	MinSet       bool
	MinExclusive bool

	Max          int64
	MaxSet       bool
	MaxExclusive bool
}

func (t *Int) SetMultipleOf(v int) {
	t.MultipleOfSet = true
	t.MultipleOf = v
}

func (t *Int) SetExclusiveMinimum(v int64) {
	t.MinExclusive = true
	t.SetMinimum(v)
}

func (t *Int) SetExclusiveMaximum(v int64) {
	t.MaxExclusive = true
	t.SetMaximum(v)
}

func (t *Int) SetMaximum(v int64) {
	t.Max = v
	t.MaxSet = true
}

func (t *Int) SetMinimum(v int64) {
	t.Max = v
	t.MaxSet = true
}

// Validate returns error if v does not match validation rules.
func (t Int) Validate(v int64) error {
	if t.MinSet && (v < t.Min || t.MinExclusive && v == t.Min) {
		return fmt.Errorf("value %d less than %d", v, t.Min)
	}
	if t.MaxSet && (v > t.Max || t.MaxExclusive && v == t.Max) {
		return fmt.Errorf("value %d greater than %d", v, t.Min)
	}
	if t.MultipleOfSet && (v%int64(t.MultipleOf)) != 0 {
		return fmt.Errorf("%d is not multiple of %d", v, t.MultipleOf)
	}

	return nil
}

func (t Int) Set() bool {
	return t.MinSet || t.MaxSet || t.MultipleOfSet
}
