package ogenregex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	type testCase struct {
		input    string
		wantType Regexp
		wantErr  bool
	}

	tests := []testCase{
		// Conversion is not required.
		{`\0`, goRegexp{}, false},
		{`\x20`, goRegexp{}, false},
		{`\v`, goRegexp{}, false},
		{`\t`, goRegexp{}, false},
		{`\n`, goRegexp{}, false},
		{`\d`, goRegexp{}, false},
		{`\w`, goRegexp{}, false},
		{`\w{1}`, goRegexp{}, false},
		{`\w{1,}`, goRegexp{}, false},
		{`\w{1,2}`, goRegexp{}, false},
		{`\b`, goRegexp{}, false},
		{`\B`, goRegexp{}, false},
		{`\.`, goRegexp{}, false},
		{`\[`, goRegexp{}, false},
		{`\]`, goRegexp{}, false},
		{`\(`, goRegexp{}, false},
		{`\)`, goRegexp{}, false},
		{`\{`, goRegexp{}, false},
		{`\}`, goRegexp{}, false},
		{`\\`, goRegexp{}, false},
		{`\$`, goRegexp{}, false},

		// Simplification.
		{`\u000a`, goRegexp{}, false},
		{`\u{000a}`, goRegexp{}, false},
		// "\z" just unnecessarily escapes the 'z'.
		{`\z`, goRegexp{}, false},

		// Conversion is required.
		//
		// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Regular_Expressions/Character_Classes#types.
		//
		// In ECMA-262, \c + [a-fA-F] is a control character.
		{`\ca`, goRegexp{}, false},
		{`\cA`, goRegexp{}, false},
		{`\cb`, goRegexp{}, false},
		{`\cB`, goRegexp{}, false},
		// In ECMA-262, \b in a character class is a backspace.
		{`[\b]`, goRegexp{}, false},
		// ECMA-262 dot matches any single character except line terminators: \n, \r, \u2028 or \u2029.
		{`.*`, goRegexp{}, false},
		// Whitespace characters in ECMA-262 differ from those in RE2.
		//
		// Whitespace characters in ECMA-262:
		// [ \f\n\r\t\v\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000\ufeff]
		{`\s`, goRegexp{}, false},
		{`\S`, goRegexp{}, false},
		{`[\s]`, goRegexp{}, false},
		// Unicode character class escape ,\p{...}/\P{...}, in ECMA-262
		{`\p{L}`, goRegexp{}, false},
		{`\P{N}`, goRegexp{}, false},

		// Use regexp2.
		{`^(?!examples/)`, regexp2Regexp{}, false},

		// Error.
		{")", nil, true},
		{"(?`)", nil, true},
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
			a.Equal(tt.input, got.String())
		})
	}
}

func TestConvert(t *testing.T) {
	type testCase struct {
		input      string
		wantString string
		wantOk     bool
	}

	tests := []testCase{
		// Conversion is not required.
		{`\0`, `\0`, true},
		{`\x20`, `\x20`, true},
		{`\v`, `\v`, true},
		{`\t`, `\t`, true},
		{`\n`, `\n`, true},
		{`\d`, `\d`, true},
		{`\w`, `\w`, true},
		{`\w{1}`, `\w{1}`, true},
		{`\w{1,}`, `\w{1,}`, true},
		{`\w{1,2}`, `\w{1,2}`, true},
		{`\b`, `\b`, true},
		{`\B`, `\B`, true},
		{`\.`, `\.`, true},
		{`\[`, `\[`, true},
		{`\]`, `\]`, true},
		{`\(`, `\(`, true},
		{`\)`, `\)`, true},
		{`\{`, `\{`, true},
		{`\}`, `\}`, true},
		{`\\`, `\\`, true},
		{`\$`, `\$`, true},

		// Simplification.
		{`\u000a`, `\x{000a}`, true},
		{`\u{000a}`, `\x{000a}`, true},
		// "\z" just unnecessarily escapes the 'z'.
		{`\z`, `z`, true},

		// Conversion is required.
		//
		// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Regular_Expressions/Character_Classes#types.
		//
		// In ECMA-262, \c + [a-fA-F] is a control character.
		{`\ca`, `\x01`, true},
		{`\cA`, `\x01`, true},
		{`\cb`, `\x02`, true},
		{`\cB`, `\x02`, true},
		// In ECMA-262, \b in a character class is a backspace.
		{`[\b]`, `[\x08]`, true},
		// ECMA-262 dot matches any single character except line terminators: \n, \r, \u2028 or \u2029.
		{`.*`, re2Dot + `*`, true},
		// Whitespace characters in ECMA-262 differ from those in RE2.
		//
		// Whitespace characters in ECMA-262:
		// [ \f\n\r\t\v\u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000\ufeff]
		{`\s`, `[` + whitespaceChars + `]`, true},
		{`\S`, `[^` + whitespaceChars + `]`, true},
		{`[\s]`, `[` + whitespaceChars + `]`, true},

		// Use regexp2.
		{`^(?!examples/)`, ``, false},

		// Error.
		{")", ``, false},
		{"(?`)", ``, false},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			got, ok := Convert(tt.input)
			a.Equal(tt.wantOk, ok, "%q", tt.input)
			a.Equal(tt.wantString, got)
		})
	}
}
