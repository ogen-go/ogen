package naming

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRule(t *testing.T) {
	a := require.New(t)
	for _, rule := range rules {
		testFind := func(key string) {
			v, ok := Rule(key)
			a.True(ok)
			a.Equal(rule, v)
		}
		testFind(rule)
		testFind(strings.ToLower(rule))
		testFind(strings.ToUpper(rule))
		testFind(strings.ToLower(rule[:1]) + rule[1:])
	}
}

func BenchmarkRule(b *testing.B) {
	suite := [...]string{
		"wifi",
		"WiFi",
		"ASCII",
		"mp3",
		"Oauth",
		"WebP",
		"JPEG",
	}

	b.Run("Rule", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		var (
			v  string
			ok bool
		)
		for i := 0; i < b.N; i++ {
			rule := suite[i%len(suite)]
			v, ok = Rule(rule)
		}
		if ok && v == "" {
			b.Fatal("sink is empty")
		}
	})

	linear := func(s string) (string, bool) {
		for _, rule := range &rules {
			if strings.EqualFold(s, rule) {
				return rule, true
			}
		}
		return "", false
	}
	b.Run("LinearSearch", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		var (
			v  string
			ok bool
		)
		for i := 0; i < b.N; i++ {
			rule := suite[i%len(suite)]
			v, ok = linear(rule)
		}
		if ok && v == "" {
			b.Fatal("sink is empty")
		}
	})
}
