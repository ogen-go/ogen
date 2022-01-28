package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt_Set(t *testing.T) {
	assert.False(t, Int{}.Set())
	assert.True(t, Int{MaxSet: true}.Set())
	assert.True(t, Int{MinSet: true}.Set())
	assert.True(t, Int{MultipleOfSet: true}.Set())
}

func TestInt_Setters(t *testing.T) {
	for _, tc := range []struct {
		do       func(*Int)
		expected Int
	}{
		{
			do: func(i *Int) {
				i.SetMultipleOf(10)
			},
			expected: Int{
				MultipleOf:    10,
				MultipleOfSet: true,
			},
		},
		{
			do: func(i *Int) {
				i.SetExclusiveMaximum(10)
			},
			expected: Int{
				Max:          10,
				MaxExclusive: true,
				MaxSet:       true,
			},
		},
		{
			do: func(i *Int) {
				i.SetExclusiveMinimum(10)
			},
			expected: Int{
				Min:          10,
				MinExclusive: true,
				MinSet:       true,
			},
		},
		{
			do: func(i *Int) {
				i.SetMaximum(10)
			},
			expected: Int{
				Max:    10,
				MaxSet: true,
			},
		},
		{
			do: func(i *Int) {
				i.SetMinimum(10)
			},
			expected: Int{
				Min:    10,
				MinSet: true,
			},
		},
	} {
		var r Int
		tc.do(&r)
		assert.Equal(t, tc.expected, r)
	}
}

func TestInt_Validate(t *testing.T) {
	for _, tc := range []struct {
		Name      string
		Validator Int
		Value     int64
		Valid     bool
	}{
		{Name: "Zero", Valid: true},
		{
			Name:      "MaxOk",
			Validator: Int{Max: 10, MaxSet: true},
			Value:     10,
			Valid:     true,
		},
		{
			Name:      "MaxErr",
			Validator: Int{Max: 10, MaxSet: true},
			Value:     11,
			Valid:     false,
		},
		{
			Name:      "MaxExclErr",
			Validator: Int{Max: 10, MaxSet: true, MaxExclusive: true},
			Value:     10,
			Valid:     false,
		},
		{
			Name:      "MinOk",
			Validator: Int{Min: 10, MinSet: true},
			Value:     10,
			Valid:     true,
		},
		{
			Name:      "MinErr",
			Validator: Int{Min: 10, MinSet: true},
			Value:     9,
			Valid:     false,
		},
		{
			Name:      "MinExclErr",
			Validator: Int{Min: 10, MinSet: true, MinExclusive: true},
			Value:     10,
			Valid:     false,
		},
		{
			Name:      "MultipleOfOk",
			Validator: Int{MultipleOf: 2, MultipleOfSet: true},
			Value:     8,
			Valid:     true,
		},
		{
			Name:      "MultipleOfErr",
			Validator: Int{MultipleOf: 2, MultipleOfSet: true},
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
