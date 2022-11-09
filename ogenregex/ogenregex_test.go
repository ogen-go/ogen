package ogenregex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	type testCase struct {
		input      string
		wantType   Regexp
		wantString string
		wantErr    bool
	}

	tests := []testCase{
		// Conversion is not required.
		{`\x20`, goRegexp{}, `\x20`, false},
		{`\v`, goRegexp{}, `\v`, false},
		{`\t`, goRegexp{}, `\t`, false},
		{`\n`, goRegexp{}, `\n`, false},
		{`\d`, goRegexp{}, `\d`, false},
		{`\w`, goRegexp{}, `\w`, false},
		{`\w{1}`, goRegexp{}, `\w{1}`, false},
		{`\w{1,}`, goRegexp{}, `\w{1,}`, false},
		{`\w{1,2}`, goRegexp{}, `\w{1,2}`, false},
		{`\b`, goRegexp{}, `\b`, false},
		{`\B`, goRegexp{}, `\B`, false},
		{`\.`, goRegexp{}, `\.`, false},
		{`\[`, goRegexp{}, `\[`, false},
		{`\]`, goRegexp{}, `\]`, false},
		{`\(`, goRegexp{}, `\(`, false},
		{`\)`, goRegexp{}, `\)`, false},
		{`\{`, goRegexp{}, `\{`, false},
		{`\}`, goRegexp{}, `\}`, false},
		{`\\`, goRegexp{}, `\\`, false},
		{`\$`, goRegexp{}, `\$`, false},

		// Simplification.
		{`\u000a`, goRegexp{}, `\x{000a}`, false},
		{`\u{000a}`, goRegexp{}, `\x{000a}`, false},
		// "\z" just unnecessarily escapes the 'z'.
		{`\z`, goRegexp{}, `z`, false},

		// Conversion is required.
		//
		// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Regular_Expressions/Character_Classes#types.
		//
		// In ECMA-262, \c + [a-fA-F] is a control character.
		{`\ca`, goRegexp{}, `\x01`, false},
		{`\cA`, goRegexp{}, `\x01`, false},
		{`\cb`, goRegexp{}, `\x02`, false},
		{`\cB`, goRegexp{}, `\x02`, false},
		// In ECMA-262, \b in a character class is a backspace.
		{`[\b]`, goRegexp{}, `[\x08]`, false},
		// ECMA-262 dot matches any single character except line terminators: \n, \r, \u2028 or \u2029.
		{`.*`, goRegexp{}, re2Dot + `*`, false},
		// Whitespace characters in ECMA-262 differ from those in RE2.
		//
		// Whitespace characters in ECMA-262:
		// [ \f\n\r\t\v\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000\ufeff]
		{`\s`, goRegexp{}, `[` + whitespaceChars + `]`, false},
		{`\S`, goRegexp{}, `[^` + whitespaceChars + `]`, false},
		{`[\s]`, goRegexp{}, `[` + whitespaceChars + `]`, false},

		// Use regexp2.
		{`^(?!examples/)`, regexp2Regexp{}, `^(?!examples/)`, false},

		// Error.
		{")", nil, ``, true},
		{"(?`)", nil, ``, true},
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
