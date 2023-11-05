package validate

import (
	"github.com/go-faster/errors"
)

// Array validates array length.
type Array struct {
	MinLength    int
	MinLengthSet bool
	MaxLength    int
	MaxLengthSet bool
	UniqueItems  bool
}

// SetMaxLength sets MaxLength validation.
func (t *Array) SetMaxLength(v int) {
	t.MaxLengthSet = true
	t.MaxLength = v
}

// SetMinLength sets MinLength validation.
func (t *Array) SetMinLength(v int) {
	t.MinLengthSet = true
	t.MinLength = v
}

// SetUniqueItems sets UniqueItems validation.
func (t *Array) SetUniqueItems(v bool) {
	t.UniqueItems = v
}

// Set reports whether any validations are set.
func (t Array) Set() bool {
	return t.MaxLengthSet || t.MinLengthSet || t.UniqueItems
}

// ValidateLength returns error if array length v is invalid.
func (t Array) ValidateLength(v int) error {
	if t.MaxLengthSet && v > t.MaxLength {
		return &MaxLengthError{Len: v, MaxLength: t.MaxLength}
	}
	if t.MinLengthSet && v < t.MinLength {
		return &MinLengthError{Len: v, MinLength: t.MinLength}
	}

	return nil
}

// UniqueItems ensures given array has no duplicates.
func UniqueItems[S ~[]T, T comparable](arr S) error {
	if len(arr) < 2 {
		return nil
	}
	for i, a := range arr {
		for _, b := range arr[i+1:] {
			if a == b {
				return errors.Errorf("duplicate element [%d] %v", i, a)
			}
		}
	}
	return nil
}
