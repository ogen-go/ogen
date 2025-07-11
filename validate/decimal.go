package validate

import (
	"github.com/go-faster/errors"
	"github.com/shopspring/decimal"
)

// Decimal validates decimal numbers.
type Decimal struct {
	MultipleOf    decimal.Decimal
	MultipleOfSet bool

	Min          decimal.Decimal
	MinSet       bool
	MinExclusive bool

	Max          decimal.Decimal
	MaxSet       bool
	MaxExclusive bool
}

// SetMultipleOf sets multipleOf validator.
func (t *Decimal) SetMultipleOf(d decimal.Decimal) {
	t.MultipleOfSet = true
	t.MultipleOf = d
}

// SetExclusiveMinimum sets exclusive minimum value.
func (t *Decimal) SetExclusiveMinimum(v decimal.Decimal) {
	t.MinExclusive = true
	t.SetMinimum(v)
}

// SetExclusiveMaximum sets exclusive maximum value.
func (t *Decimal) SetExclusiveMaximum(v decimal.Decimal) {
	t.MaxExclusive = true
	t.SetMaximum(v)
}

// SetMinimum sets minimum value.
func (t *Decimal) SetMinimum(v decimal.Decimal) {
	t.Min = v
	t.MinSet = true
}

// SetMaximum sets maximum value.
func (t *Decimal) SetMaximum(v decimal.Decimal) {
	t.Max = v
	t.MaxSet = true
}

// Set reports whether any validations are set.
func (t Decimal) Set() bool {
	return t.MinSet || t.MaxSet || t.MultipleOfSet
}

// Validate returns error if v does not match validation rules.
func (t Decimal) Validate(v decimal.Decimal) error {
	return t.validate(v)
}

func (t Decimal) validate(v decimal.Decimal) error {
	if t.MinSet {
		cmp := v.Cmp(t.Min)
		if cmp < 0 || (t.MinExclusive && cmp == 0) {
			return errors.Errorf("value %s less than %s", v.String(), t.Min.String())
		}
	}
	if t.MaxSet {
		cmp := v.Cmp(t.Max)
		if cmp > 0 || (t.MaxExclusive && cmp == 0) {
			return errors.Errorf("value %s greater than %s", v.String(), t.Max.String())
		}
	}
	if t.MultipleOfSet {
		if !v.Mod(t.MultipleOf).IsZero() {
			return errors.Errorf("value %s is not multiple of %s", v.String(), t.MultipleOf.String())
		}
	}

	return nil
}
