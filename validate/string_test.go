package validate

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmail(t *testing.T) {
	v := String{Email: true}
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
		"hello",
		"@",
		"@test",
		"test@",
	} {
		require.Error(t, v.Validate(s), "%q should be invalid", s)
	}
}

func TestRegex(t *testing.T) {
	v := String{Regex: regexp.MustCompile(`^\d$`)}
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
