package json

const MaxMoreLevel = 10

// More is helper for writing commas.
//
// Up to MaxMoreLevel levels.
type More struct {
	idx  int
	w    *Encoder
	more [MaxMoreLevel]bool
}

func (f *More) Reset() {
	f.w = nil
	f.more = [MaxMoreLevel]bool{}
}
func NewMore(w *Encoder) More { return More{w: w} }

func (f *More) Down() { f.idx++ }

func (f *More) Up() {
	f.more[f.idx] = false
	f.idx--
}

// More writes "more" (comma) if required and maintains state.
func (f *More) More() {
	if f.more[f.idx] {
		f.w.More()
	}
	f.more[f.idx] = true
}
