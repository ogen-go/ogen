// Package bitset implements a byte-slice based bitset
package bitset

// Bitset is a byte-slice based bitset.
type Bitset []uint8

// Set sets the bit by given index.
func (r *Bitset) Set(i int, v bool) {
	maskIdx := i / 8
	for len(*r) <= maskIdx {
		*r = append(*r, 0)
	}
	bitIdx := i % 8

	set := uint8(0)
	if v {
		set = 1
	}
	(*r)[maskIdx] |= set << uint8(bitIdx)
}

// Build builds a bitset from slice using given predicate.
func Build[T any](s []T, cb func(int, T) bool) (r Bitset) {
	i := 0
	r = append(r, 0)
	for elemIdx, elem := range s {
		maskIdx := i / 8
		if len(r) <= maskIdx {
			r = append(r, 0)
		}
		bitIdx := i % 8

		set := uint8(0)
		if cb(elemIdx, elem) {
			set = 1
		}
		r[maskIdx] |= set << uint8(bitIdx)
		i++
	}
	return r
}
