package validate

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat_Set(t *testing.T) {
	assert.False(t, Float{}.Set())
	assert.True(t, Float{MaxSet: true}.Set())
	assert.True(t, Float{MinSet: true}.Set())
	assert.True(t, Float{MultipleOfSet: true}.Set())
}

func TestFloat_Setters(t *testing.T) {
	for _, tc := range []struct {
		do       func(*Float)
		expected Float
	}{
		{
			do: func(i *Float) {
				i.SetMultipleOf(new(big.Rat).SetInt64(10))
			},
			expected: Float{
				MultipleOf:    new(big.Rat).SetInt64(10),
				MultipleOfSet: true,
			},
		},
		{
			do: func(i *Float) {
				i.SetExclusiveMaximum(10)
			},
			expected: Float{
				Max:          10,
				MaxExclusive: true,
				MaxSet:       true,
			},
		},
		{
			do: func(i *Float) {
				i.SetExclusiveMinimum(10)
			},
			expected: Float{
				Min:          10,
				MinExclusive: true,
				MinSet:       true,
			},
		},
		{
			do: func(i *Float) {
				i.SetMaximum(10)
			},
			expected: Float{
				Max:    10,
				MaxSet: true,
			},
		},
		{
			do: func(i *Float) {
				i.SetMinimum(10)
			},
			expected: Float{
				Min:    10,
				MinSet: true,
			},
		},
	} {
		var r Float
		tc.do(&r)
		assert.Equal(t, tc.expected, r)
	}
}

func TestFloat_Validate(t *testing.T) {
	for _, tc := range []struct {
		Name      string
		Validator Float
		Value     float64
		Valid     bool
	}{
		{Name: "Zero", Valid: true},
		{
			Name:      "MaxOk",
			Validator: Float{Max: 10, MaxSet: true},
			Value:     10,
			Valid:     true,
		},
		{
			Name:      "MaxErr",
			Validator: Float{Max: 10, MaxSet: true},
			Value:     11,
			Valid:     false,
		},
		{
			Name:      "MaxExclErr",
			Validator: Float{Max: 10, MaxSet: true, MaxExclusive: true},
			Value:     10,
			Valid:     false,
		},
		{
			Name:      "MinOk",
			Validator: Float{Min: 10, MinSet: true},
			Value:     10,
			Valid:     true,
		},
		{
			Name:      "MinErr",
			Validator: Float{Min: 10, MinSet: true},
			Value:     9,
			Valid:     false,
		},
		{
			Name:      "MinExclErr",
			Validator: Float{Min: 10, MinSet: true, MinExclusive: true},
			Value:     10,
			Valid:     false,
		},
		{
			Name:      "MultipleOfOk",
			Validator: Float{MultipleOf: new(big.Rat).SetInt64(10), MultipleOfSet: true},
			Value:     20,
			Valid:     true,
		},
		{
			Name:      "MultipleOfErr",
			Validator: Float{MultipleOf: new(big.Rat).SetInt64(10), MultipleOfSet: true},
			Value:     13,
			Valid:     false,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			valid := tc.Validator.Validate(tc.Value) == nil
			assert.Equal(t, tc.Valid, valid, "%v: %+v",
				tc.Validator,
				tc.Value,
			)
		})
	}
}
