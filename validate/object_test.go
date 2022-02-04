package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObject_Set(t *testing.T) {
	var v Object
	v.SetMaxProperties(10)
	v.SetMinProperties(5)
	require.True(t, v.Set())
	require.Equal(t, Object{
		MinProperties:    5,
		MinPropertiesSet: true,
		MaxProperties:    10,
		MaxPropertiesSet: true,
	}, v)
	require.NoError(t, v.ValidateProperties(7))
}

func TestObject_ValidateProperties(t *testing.T) {
	for _, tc := range []struct {
		Name      string
		Validator Object
		Value     int
		Valid     bool
	}{
		{Name: "Zero", Valid: true},
		{
			Name:      "MaxPropertiesOk",
			Validator: Object{MaxProperties: 10, MaxPropertiesSet: true},
			Value:     5,
			Valid:     true,
		},
		{
			Name:      "MaxPropertiesErr",
			Validator: Object{MaxProperties: 10, MaxPropertiesSet: true},
			Value:     15,
			Valid:     false,
		},
		{
			Name:      "MinPropertiesOk",
			Validator: Object{MinProperties: 10, MinPropertiesSet: true},
			Value:     15,
			Valid:     true,
		},
		{
			Name:      "MinPropertiesErr",
			Validator: Object{MinProperties: 10, MinPropertiesSet: true},
			Value:     5,
			Valid:     false,
		},
		{
			Name: "BothOk",
			Validator: Object{
				MinProperties:    10,
				MinPropertiesSet: true,
				MaxProperties:    15,
				MaxPropertiesSet: true,
			},
			Value: 12,
			Valid: true,
		},
		{
			Name: "BothErrMax",
			Validator: Object{
				MinProperties:    10,
				MinPropertiesSet: true,
				MaxProperties:    15,
				MaxPropertiesSet: true,
			},
			Value: 17,
			Valid: false,
		},
		{
			Name: "BothErrMin",
			Validator: Object{
				MinProperties:    10,
				MinPropertiesSet: true,
				MaxProperties:    15,
				MaxPropertiesSet: true,
			},
			Value: 7,
			Valid: false,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			valid := tc.Validator.ValidateProperties(tc.Value) == nil
			assert.Equal(t, tc.Valid, valid, "max: %d, min: %d, v: %d",
				tc.Validator.MaxProperties,
				tc.Validator.MinProperties,
				tc.Value,
			)
		})
	}
}
