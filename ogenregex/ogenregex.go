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
	"fmt"
	"regexp"

	"github.com/dlclark/regexp2"
	"github.com/go-faster/errors"
)

var _ = []Regexp{
	goRegexp{},
	regexp2Regexp{},
}

type goRegexp struct {
	exp *regexp.Regexp
}

func (r goRegexp) Match(s []byte) (bool, error) {
	return r.exp.Match(s), nil
}

func (r goRegexp) MatchString(s string) (bool, error) {
	return r.exp.MatchString(s), nil
}

func (r goRegexp) String() string {
	return r.exp.String()
}

type regexp2Regexp struct {
	exp *regexp2.Regexp
}

func (r regexp2Regexp) Match(s []byte) (bool, error) {
	return r.exp.MatchRunes([]rune(string(s)))
}

func (r regexp2Regexp) MatchString(s string) (bool, error) {
	return r.exp.MatchString(s)
}

func (r regexp2Regexp) String() string {
	return r.exp.String()
}

// Regexp is a regular expression interface.
type Regexp interface {
	Match(s []byte) (bool, error)
	MatchString(s string) (bool, error)
	fmt.Stringer
}

// Compile compiles a regular expression.
func Compile(exp string) (Regexp, error) {
	if re, err := regexp.Compile(exp); err == nil {
		return goRegexp{re}, nil
	}
	re, err := regexp2.Compile(exp, regexp2.ECMAScript|regexp2.Unicode)
	if err != nil {
		return nil, errors.Wrap(err, "regexp2")
	}
	return regexp2Regexp{re}, nil
}

// MustCompile compiles a regular expression and panics on error.
func MustCompile(exp string) Regexp {
	return errors.Must(Compile(exp))
}
