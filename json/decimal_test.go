package json

import (
	"testing"

	"github.com/go-faster/jx"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestDecimal(t *testing.T) {
	tests := []struct {
		name    string
		value   decimal.Decimal
		wantStr string
	}{
		{
			name:    "Integer",
			value:   decimal.NewFromInt(123),
			wantStr: "123",
		},
		{
			name:    "Float",
			value:   decimal.NewFromFloat(123.456),
			wantStr: "123.456",
		},
		{
			name:    "Zero",
			value:   decimal.Zero,
			wantStr: "0",
		},
		{
			name:    "Negative",
			value:   decimal.NewFromFloat(-123.456),
			wantStr: "-123.456",
		},
	}

	t.Run("EncodeDecimal", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				e := &jx.Encoder{}
				EncodeDecimal(e, tt.value)
				require.Equal(t, tt.wantStr, string(e.Bytes()))
			})
		}
	})

	t.Run("DecodeDecimal", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				d := jx.DecodeStr(tt.wantStr)
				got, err := DecodeDecimal(d)
				require.NoError(t, err)
				require.True(t, tt.value.Equal(got), "expected %s, got %s", tt.value, got)
			})
		}
	})

	t.Run("EncodeStringDecimal", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				e := &jx.Encoder{}
				EncodeStringDecimal(e, tt.value)
				expected := `"` + tt.wantStr + `"`
				require.Equal(t, expected, string(e.Bytes()))
			})
		}
	})

	t.Run("DecodeStringDecimal", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				input := `"` + tt.wantStr + `"`
				d := jx.DecodeStr(input)
				got, err := DecodeStringDecimal(d)
				require.NoError(t, err)
				require.True(t, tt.value.Equal(got), "expected %s, got %s", tt.value, got)
			})
		}
	})
}

func TestDecimalErrors(t *testing.T) {
	t.Run("DecodeDecimal_InvalidNumber", func(t *testing.T) {
		d := jx.DecodeStr(`"not a number"`)
		_, err := DecodeDecimal(d)
		require.Error(t, err)
	})

	t.Run("DecodeStringDecimal_InvalidNumber", func(t *testing.T) {
		d := jx.DecodeStr(`"not a number"`)
		_, err := DecodeStringDecimal(d)
		require.Error(t, err)
	})
}
