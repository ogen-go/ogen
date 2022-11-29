package ogenregex

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/stretchr/testify/require"
)

type cache[R fmt.Stringer] map[string]R

func (c cache[R]) Get(re string, compile func(re string) (R, error)) (R, error) {
	if r, ok := c[re]; ok {
		return r, nil
	}
	r, err := compile(re)
	if err != nil {
		var zero R
		return zero, err
	}
	c[r.String()] = r
	return r, nil
}

func compileDlclark(re string) (*regexp2.Regexp, error) {
	r, err := regexp2.Compile(re, regexp2Flags)
	if err != nil {
		return nil, err
	}
	r.MatchTimeout = 15 * time.Second
	return r, nil
}

func FuzzConvertMatch(f *testing.F) {
	for _, seed := range []struct {
		re, input string
	}{
		{`^a$`, `a`},
		{`^\x20$`, "\x20"},

		{`^\w{2,3}$`, "a"},
		{`^\w{2,3}$`, "ab"},

		{`^\ca$`, "a"},
		{`^\ca$`, "\x01"},

		{`[\b]+`, ""},
		{`[\b]+`, "\x08"},
		{`[\b]+`, "\x08\x08"},
		{`[\b]+`, "abc"},

		{`\s+`, "\x20"},
		{`\s+`, "\n"},
		{`\s+`, "\t"},
		{`\s+`, "\u2028"},
		{`\s+`, "abc"},

		{`.*`, ""},
		{`.*`, "\n"},
		{`.*`, "\u2029"},
		{`.*`, "a"},
	} {
		f.Add(seed.re, seed.input)
	}

	var (
		goCache      = cache[*regexp.Regexp]{}
		dlclarkCache = cache[*regexp2.Regexp]{}
	)

	f.Fuzz(func(t *testing.T, re, input string) {
		var converted string
		defer func() {
			if r := recover(); t.Skipped() || t.Failed() || r != nil {
				t.Logf("Regexp: %q", re)
				t.Logf("Input: %q", input)
				t.Logf("Converted: %q", converted)
			}
		}()

		converted, ok := Convert(re)
		if !ok {
			t.Skip("Can't convert")
			return
		}
		a := require.New(t)

		goCompiled, err := goCache.Get(converted, regexp.Compile)
		a.NoError(err)

		dlclarkCompiled, err := dlclarkCache.Get(re, compileDlclark)
		a.NoError(err)

		expected, err := dlclarkCompiled.MatchString(input)
		if err != nil {
			t.Skipf("Can't match: %v", err)
			return
		}

		got := goCompiled.MatchString(input)
		a.Equal(expected, got)
	})
}
