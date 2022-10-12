package validate

import "github.com/go-faster/errors"

// Int validates integers.
type Int struct {
	MultipleOf    uint64
	MultipleOfSet bool

	Min          int64
	MinSet       bool
	MinExclusive bool

	Max          int64
	MaxSet       bool
	MaxExclusive bool
}

// SetMultipleOf sets multipleOf validator.
func (t *Int) SetMultipleOf(v uint64) {
	t.MultipleOfSet = true
	t.MultipleOf = v
}

// SetExclusiveMinimum sets exclusive minimum value.
func (t *Int) SetExclusiveMinimum(v int64) {
	t.MinExclusive = true
	t.SetMinimum(v)
}

// SetExclusiveMaximum sets exclusive maximum value.
func (t *Int) SetExclusiveMaximum(v int64) {
	t.MaxExclusive = true
	t.SetMaximum(v)
}

// SetMinimum sets minimum value.
func (t *Int) SetMinimum(v int64) {
	t.Min = v
	t.MinSet = true
}

// SetMaximum sets maximum value.
func (t *Int) SetMaximum(v int64) {
	t.Max = v
	t.MaxSet = true
}

// Set reports whether any validations are set.
func (t Int) Set() bool {
	return t.MinSet || t.MaxSet || t.MultipleOfSet
}

// Validate returns error if v does not match validation rules.
func (t Int) Validate(v int64) error {
	if t.MinSet && (v < t.Min || t.MinExclusive && v == t.Min) {
		return errors.Errorf("value %d less than %d", v, t.Min)
	}
	if t.MaxSet && (v > t.Max || t.MaxExclusive && v == t.Max) {
		return errors.Errorf("value %d greater than %d", v, t.Max)
	}
	// We don't care about sign when checking value using multipleOf.
	if v < 0 {
		v *= -1
	}
	if t.MultipleOfSet && (uint64(v)%t.MultipleOf) != 0 {
		return errors.Errorf("value %d is not multiple of %d", v, t.MultipleOf)
	}

	return nil
}
