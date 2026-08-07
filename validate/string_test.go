package validate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/ogenregex"
)

func TestEmail(t *testing.T) {
	v := String{Email: true}
	require.True(t, v.Set())

	for _, s := range []string{
		"foo@example",
		"foo@example.com",
		"foo@казахстан",
		// Quoted local parts may legally contain spaces, '@' and other
		// otherwise-disallowed characters. See #1419.
		`"foo bar"@example.com`,
		`"foo@bar"@example.com`,
		`"@"@example.com`,
	} {
		require.NoError(t, v.Validate(s))
	}
	for _, s := range []string{
		"foo @example", // unquoted space
		"",
		"\x00",   // not printable
		"\n",     // space character
		"\u202f", // unicode space character
		"hello",
		"@",
		"@@",
		"@test",
		"a@@test", // unquoted, multiple @
		"test@",
	} {
		require.Error(t, v.Validate(s), "%q should be invalid", s)
	}
}

func TestHostname(t *testing.T) {
	v := String{Hostname: true}
	require.True(t, v.Set())

	for _, s := range []string{
		"example.com",
		"foo",
		"bar-baz.ch",
	} {
		require.NoError(t, v.Validate(s))
	}
	for _, s := range []string{
		"",
		"\x00",                   // not printable
		"\n",                     // space character
		"\u202f",                 // unicode space character
		strings.Repeat("a", 257), // too long
		"Щ",                      // non-ASCII
		"@",
	} {
		require.Error(t, v.Validate(s), "%q should be invalid", s)
	}
}

func TestRegex(t *testing.T) {
	v := String{Regex: ogenregex.MustCompile(`^\d$`)}
	require.True(t, v.Set())

	for _, s := range []string{
		"1",
		"2",
	} {
		require.NoError(t, v.Validate(s))
	}
	for _, s := range []string{
		"s10",
		"",
		"hello",
	} {
		require.Error(t, v.Validate(s), "%q should be invalid", s)
	}
}

func TestString_Validate(t *testing.T) {
	v := String{}
	require.False(t, v.Set())

	v.SetMinLength(2)
	require.True(t, v.Set())

	v.SetMaxLength(5)
	require.True(t, v.Set())

	for _, s := range []string{
		"123",
		"abc",
		"щщщщ",
	} {
		require.NoError(t, v.Validate(s))
	}
	for _, s := range []string{
		"",
		"s",
		"щ",
		"щщщщщщ",
		"ssssss",
	} {
		require.Error(t, v.Validate(s), "%q should be invalid", s)
	}
}

func TestString_ValidateNumeric(t *testing.T) {
	minMax := String{MinNumeric: 1, MinNumericSet: true, MaxNumeric: 10, MaxNumericSet: true}
	minOnly := String{MinNumeric: 1, MinNumericSet: true}
	maxOnly := String{MaxNumeric: 10, MaxNumericSet: true}

	require.True(t, minOnly.Set())
	require.True(t, maxOnly.Set())

	for _, tc := range []struct {
		Name      string
		Validator String
		Value     string
		Valid     bool
	}{
		{Name: "Min", Validator: minMax, Value: "1", Valid: true},
		{Name: "Max", Validator: minMax, Value: "10", Valid: true},
		{Name: "Fraction", Validator: minMax, Value: "5.5", Valid: true},
		{Name: "Sign", Validator: minMax, Value: "+5", Valid: true},
		{Name: "LeadingZeroes", Validator: minMax, Value: "007", Valid: true},
		{Name: "Exponent", Validator: minMax, Value: "1e1", Valid: true},
		{Name: "BelowMin", Validator: minMax, Value: "0", Valid: false},
		{Name: "BelowMinFraction", Validator: minMax, Value: "0.5", Valid: false},
		{Name: "AboveMax", Validator: minMax, Value: "11", Valid: false},
		{Name: "Empty", Validator: minMax, Value: "", Valid: false},
		{Name: "NotANumber", Validator: minMax, Value: "abc", Valid: false},
		{Name: "TrailingGarbage", Validator: minMax, Value: "5abc", Valid: false},
		{Name: "GroupSeparator", Validator: minMax, Value: "5,000", Valid: false},
		{Name: "SecondDot", Validator: minMax, Value: "1.2.3", Valid: false},
		{Name: "LeadingSpace", Validator: minMax, Value: " 5", Valid: false},
		{Name: "TrailingSpace", Validator: minMax, Value: "5 ", Valid: false},
		{Name: "NaN", Validator: minMax, Value: "NaN", Valid: false},
		{Name: "NaNLower", Validator: minMax, Value: "nan", Valid: false},
		{Name: "PosInfMinOnly", Validator: minOnly, Value: "Inf", Valid: false},
		{Name: "PosInfinityMinOnly", Validator: minOnly, Value: "Infinity", Valid: false},
		{Name: "NaNMinOnly", Validator: minOnly, Value: "NaN", Valid: false},
		{Name: "NegInfMaxOnly", Validator: maxOnly, Value: "-Inf", Valid: false},
		{Name: "NaNMaxOnly", Validator: maxOnly, Value: "NaN", Valid: false},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			valid := tc.Validator.Validate(tc.Value) == nil
			require.Equal(t, tc.Valid, valid, "%+v: %q", tc.Validator, tc.Value)
		})
	}
}
