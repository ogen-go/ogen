package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArray_Set(t *testing.T) {
	for _, tc := range []struct {
		Name      string
		Validator Array
		Value     int
		Valid     bool
	}{
		{Name: "Zero", Valid: true},
		{
			Name:      "MaxLengthOk",
			Validator: Array{MaxLength: 10, MaxLengthSet: true},
			Value:     5,
			Valid:     true,
		},
		{
			Name:      "MaxLengthErr",
			Validator: Array{MaxLength: 10, MaxLengthSet: true},
			Value:     15,
			Valid:     false,
		},
		{
			Name:      "MinLengthOk",
			Validator: Array{MinLength: 10, MinLengthSet: true},
			Value:     15,
			Valid:     true,
		},
		{
			Name:      "MinLengthErr",
			Validator: Array{MinLength: 10, MinLengthSet: true},
			Value:     5,
			Valid:     false,
		},
		{
			Name: "BothOk",
			Validator: Array{
				MinLength:    10,
				MinLengthSet: true,
				MaxLength:    15,
				MaxLengthSet: true,
			},
			Value: 12,
			Valid: true,
		},
		{
			Name: "BothErrMax",
			Validator: Array{
				MinLength:    10,
				MinLengthSet: true,
				MaxLength:    15,
				MaxLengthSet: true,
			},
			Value: 17,
			Valid: false,
		},
		{
			Name: "BothErrMin",
			Validator: Array{
				MinLength:    10,
				MinLengthSet: true,
				MaxLength:    15,
				MaxLengthSet: true,
			},
			Value: 7,
			Valid: false,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			valid := tc.Validator.ValidateLength(tc.Value) == nil
			assert.Equal(t, tc.Valid, valid, "max: %d, min: %d, v: %d",
				tc.Validator.MaxLength,
				tc.Validator.MinLength,
				tc.Value,
			)
		})
	}
}
