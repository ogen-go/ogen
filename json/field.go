package json

const MaxMoreLevel = 10

// More is helper for writing commas.
//
// Up to MaxMoreLevel levels.
type More struct {
	idx  int
	s    *Stream
	more [MaxMoreLevel]bool
}

func (f *More) Reset() {
	f.s = nil
	f.more = [MaxMoreLevel]bool{}
}
func NewMore(s *Stream) More { return More{s: s} }

func (f *More) Down() { f.idx++ }

func (f *More) Up() {
	f.more[f.idx] = false
	f.idx--
}

// More writes "more" (comma) if required and maintans state.
func (f *More) More() {
	if f.more[f.idx] {
		f.s.WriteMore()
	}
	f.more[f.idx] = true
}
