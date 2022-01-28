package validate

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmail(t *testing.T) {
	v := String{Email: true}
	require.True(t, v.Set())

	for _, s := range []string{
		"foo@example",
		"foo@example.com",
		"foo@казахстан",
	} {
		require.NoError(t, v.Validate(s))
	}
	for _, s := range []string{
		"foo @example",
		"",
		"\x00",   // not printable
		"\n",     // space character
		"\u202f", // unicode space character
		"hello",
		"@",
		"@@",
		"@test",
		"a@@test",
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
	v := String{Regex: regexp.MustCompile(`^\d$`)}
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
