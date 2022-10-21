package xslices

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	a := require.New(t)

	odd := func(x int) bool {
		return x%2 == 1
	}
	filter := func(v []int, cb func(int) bool) []int {
		Filter(&v, cb)
		return v
	}

	a.Empty(filter([]int(nil), odd))
	a.Empty(filter([]int{}, odd))
	a.Equal(filter([]int{1, 2, 3, 4, 5}, odd), []int{1, 3, 5})
}

func TestContainsFunc(t *testing.T) {
	a := require.New(t)
	foldEqual := func(a string) func(string) bool {
		return func(b string) bool {
			return strings.EqualFold(a, b)
		}
	}
	a.False(ContainsFunc([]string(nil), foldEqual("a")))
	a.False(ContainsFunc([]string{}, foldEqual("a")))
	a.True(ContainsFunc([]string{"a"}, foldEqual("a")))
	a.True(ContainsFunc([]string{"A"}, foldEqual("a")))
	a.True(ContainsFunc([]string{"foo"}, foldEqual("FoO")))
	a.True(ContainsFunc([]string{"FoO"}, foldEqual("foo")))
}

func TestFindFunc(t *testing.T) {
	findA := func(s string) bool {
		return s == "a"
	}
	tests := []struct {
		r      []string
		f      func(string) bool
		wantR  string
		wantOk bool
	}{
		{nil, findA, "", false},
		{[]string{}, findA, "", false},
		{[]string{"b"}, findA, "", false},
		{[]string{"a", "b"}, findA, "a", true},
		{[]string{"b", "a"}, findA, "a", true},
		{[]string{"a", "a"}, findA, "a", true},
	}
	a := require.New(t)
	for _, tt := range tests {
		gotR, gotOk := FindFunc(tt.r, tt.f)
		a.Equal(tt.wantR, gotR)
		a.Equal(tt.wantOk, gotOk)
	}
}
