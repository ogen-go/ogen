// Package ogenregex provides an interface to the regex engine.
//
// JSON Schema specification prefers to use ECMA 262 regular expressions. However, Go's
// regex engine is based on RE2, which is a different engine. Also, Go's regex engine
// does not support lookbehind assertions, to ensure linear time matching.
//
// This package provides unified interface to both engines. Go's regex engine is used
// by default, but if the regex is not supported, the dlclark/regexp2 would be used.
package ogenregex

import (
	"regexp"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/go-faster/errors"
)

var _ = []Regexp{
	goRegexp{},
	regexp2Regexp{},
}

type goRegexp struct {
	re *regexp.Regexp
}

// Go returns wrapper for Go's regexp.
func Go(re *regexp.Regexp) Regexp {
	return goRegexp{re}
}

func (r goRegexp) Match(s []byte) (bool, error) {
	return r.re.Match(s), nil
}

func (r goRegexp) MatchString(s string) (bool, error) {
	return r.re.MatchString(s), nil
}

func (r goRegexp) String() string {
	return r.re.String()
}

type regexp2Regexp struct {
	re *regexp2.Regexp
}

// Regexp2 returns wrapper for dlclark/regexp2.
func Regexp2(re *regexp2.Regexp) Regexp {
	return regexp2Regexp{re}
}

func (r regexp2Regexp) Match(s []byte) (bool, error) {
	return r.re.MatchRunes([]rune(string(s)))
}

func (r regexp2Regexp) MatchString(s string) (bool, error) {
	return r.re.MatchString(s)
}

func (r regexp2Regexp) String() string {
	return r.re.String()
}

// Regexp is a regular expression interface.
type Regexp interface {
	Match(s []byte) (bool, error)
	MatchString(s string) (bool, error)
	String() string
}

// Compile compiles a regular expression.
//
// NOTE: this function may compile the same expression multiple times and can
// be slow. Call it once and reuse the result.
func Compile(exp string) (Regexp, error) {
	if converted, ok := Convert(exp); ok {
		if re, err := regexp.Compile(converted); err == nil {
			return Go(re), nil
		}
	}
	re, err := regexp2.Compile(exp, regexp2.ECMAScript|regexp2.Unicode)
	if err != nil {
		return nil, errors.Wrap(err, "regexp2")
	}
	// FIXME(tdakkota): Default timeout is "forever", which may lead to DoS.
	// 	Probably, we should make this configurable.
	re.MatchTimeout = 15 * time.Second
	return Regexp2(re), nil
}

// MustCompile compiles a regular expression and panics on error.
//
// NOTE: this function may compile the same expression multiple times and can
// be slow. Compile the expression once and reuse it.
func MustCompile(exp string) Regexp {
	return errors.Must(Compile(exp))
}
