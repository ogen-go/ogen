package validate

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDecimal_Set(t *testing.T) {
	assert.False(t, Decimal{}.Set())
	assert.True(t, Decimal{MaxSet: true}.Set())
	assert.True(t, Decimal{MinSet: true}.Set())
	assert.True(t, Decimal{MultipleOfSet: true}.Set())
}

func TestDecimal_Setters(t *testing.T) {
	for _, tc := range []struct {
		do       func(*Decimal)
		expected Decimal
	}{
		{
			do: func(i *Decimal) {
				i.SetMultipleOf(decimal.NewFromInt(10))
			},
			expected: Decimal{
				MultipleOf:    decimal.NewFromInt(10),
				MultipleOfSet: true,
			},
		},
		{
			do: func(i *Decimal) {
				i.SetExclusiveMaximum(decimal.NewFromInt(10))
			},
			expected: Decimal{
				Max:          decimal.NewFromInt(10),
				MaxExclusive: true,
				MaxSet:       true,
			},
		},
		{
			do: func(i *Decimal) {
				i.SetExclusiveMinimum(decimal.NewFromInt(10))
			},
			expected: Decimal{
				Min:          decimal.NewFromInt(10),
				MinExclusive: true,
				MinSet:       true,
			},
		},
		{
			do: func(i *Decimal) {
				i.SetMaximum(decimal.NewFromInt(10))
			},
			expected: Decimal{
				Max:    decimal.NewFromInt(10),
				MaxSet: true,
			},
		},
		{
			do: func(i *Decimal) {
				i.SetMinimum(decimal.NewFromInt(10))
			},
			expected: Decimal{
				Min:    decimal.NewFromInt(10),
				MinSet: true,
			},
		},
	} {
		var r Decimal
		tc.do(&r)
		assert.Equal(t, tc.expected, r)
	}
}

func TestDecimal_Validate(t *testing.T) {
	for _, tc := range []struct {
		Name      string
		Validator Decimal
		Value     decimal.Decimal
		Valid     bool
	}{
		{Name: "Zero", Valid: true},
		{
			Name:      "MaxOk",
			Validator: Decimal{Max: decimal.NewFromInt(10), MaxSet: true},
			Value:     decimal.NewFromInt(10),
			Valid:     true,
		},
		{
			Name:      "MaxErr",
			Validator: Decimal{Max: decimal.NewFromInt(10), MaxSet: true},
			Value:     decimal.NewFromInt(11),
			Valid:     false,
		},
		{
			Name:      "MaxExclErr",
			Validator: Decimal{Max: decimal.NewFromInt(10), MaxSet: true, MaxExclusive: true},
			Value:     decimal.NewFromInt(10),
			Valid:     false,
		},
		{
			Name:      "MinOk",
			Validator: Decimal{Min: decimal.NewFromInt(10), MinSet: true},
			Value:     decimal.NewFromInt(10),
			Valid:     true,
		},
		{
			Name:      "MinErr",
			Validator: Decimal{Min: decimal.NewFromInt(10), MinSet: true},
			Value:     decimal.NewFromInt(9),
			Valid:     false,
		},
		{
			Name:      "MinExclErr",
			Validator: Decimal{Min: decimal.NewFromInt(10), MinSet: true, MinExclusive: true},
			Value:     decimal.NewFromInt(10),
			Valid:     false,
		},
		{
			Name:      "MultipleOfOk",
			Validator: Decimal{MultipleOf: decimal.NewFromInt(10), MultipleOfSet: true},
			Value:     decimal.NewFromInt(20),
			Valid:     true,
		},
		{
			Name:      "MultipleOfErr",
			Validator: Decimal{MultipleOf: decimal.NewFromInt(10), MultipleOfSet: true},
			Value:     decimal.NewFromInt(13),
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
