package validate

import (
	"math"
	"math/big"

	"github.com/go-faster/errors"
)

// Float validates float numbers.
type Float struct {
	MultipleOf    *big.Rat
	MultipleOfSet bool

	Min          float64
	MinSet       bool
	MinExclusive bool

	Max          float64
	MaxSet       bool
	MaxExclusive bool
}

// SetMultipleOf sets multipleOf validator.
func (t *Float) SetMultipleOf(rat *big.Rat) {
	t.MultipleOfSet = true
	t.MultipleOf = rat
}

// SetExclusiveMinimum sets exclusive minimum value.
func (t *Float) SetExclusiveMinimum(v float64) {
	t.MinExclusive = true
	t.SetMinimum(v)
}

// SetExclusiveMaximum sets exclusive maximum value.
func (t *Float) SetExclusiveMaximum(v float64) {
	t.MaxExclusive = true
	t.SetMaximum(v)
}

// SetMinimum sets minimum value.
func (t *Float) SetMinimum(v float64) {
	t.Min = v
	t.MinSet = true
}

// SetMaximum sets maximum value.
func (t *Float) SetMaximum(v float64) {
	t.Max = v
	t.MaxSet = true
}

// Set reports whether any validations are set.
func (t Float) Set() bool {
	return t.MinSet || t.MaxSet || t.MultipleOfSet
}

// Validate returns error if v does not match validation rules.
func (t Float) Validate(v float64) error {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return errors.Errorf("%f float value is invalid", v)
	}
	if t.MinSet && (v < t.Min || t.MinExclusive && v == t.Min) {
		return errors.Errorf("value %f less than %f", v, t.Min)
	}
	if t.MaxSet && (v > t.Max || t.MaxExclusive && v == t.Max) {
		return errors.Errorf("value %f greater than %f", v, t.Max)
	}
	if t.MultipleOfSet {
		val := new(big.Rat).SetFloat64(v)
		if !val.Quo(val, t.MultipleOf).IsInt() {
			return errors.Errorf("%f is not multiple of %s", v, t.MultipleOf.String())
		}
	}

	return nil
}
