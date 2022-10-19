package ogenregex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	tests := []struct {
		input      string
		wantType   Regexp
		wantString string
		wantErr    bool
	}{
		{`\d`, goRegexp{}, `\d`, false},
		{`\w`, goRegexp{}, `\w`, false},
		{`^(?!examples/)`, regexp2Regexp{}, `^(?!examples/)`, false},

		{")", nil, ``, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			got, err := Compile(tt.input)
			if tt.wantErr {
				a.Error(err)
				t.Log(err.Error())
				a.Panics(func() { MustCompile(tt.input) })
				return
			}
			a.NoError(err)
			a.NotPanics(func() { MustCompile(tt.input) })
			a.IsType(tt.wantType, got)
			a.Equal(tt.wantString, got.String())
		})
	}
}
