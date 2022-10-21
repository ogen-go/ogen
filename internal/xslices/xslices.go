// Package xslices provides some generic utilities missed from x/exp/slices.
package xslices

import "golang.org/x/exp/slices"

// Filter performs in-place filtering of a slice.
func Filter[S ~[]E, E any](sptr *S, keep func(E) bool) {
	var (
		n int
		s = *sptr
	)
	for _, v := range s {
		if keep(v) {
			s[n] = v
			n++
		}
	}
	*sptr = s[:n]
}

// ContainsFunc returns true if the slice contains an element satisfying the predicate.
func ContainsFunc[S ~[]E, E any](s S, equal func(E) bool) bool {
	return slices.IndexFunc(s, equal) >= 0
}

// FindFunc returns the first element satisfying the predicate.
func FindFunc[S ~[]E, E any](s S, equal func(E) bool) (r E, _ bool) {
	idx := slices.IndexFunc(s, equal)
	if idx < 0 {
		return r, false
	}
	return s[idx], true
}
