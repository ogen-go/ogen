package json

// FieldWriter is helper for writing fields with ",".
type FieldWriter struct {
	s          *Stream
	shouldMore bool
}

func NewFieldWriter(s *Stream) FieldWriter {
	return FieldWriter{
		s: s,
	}
}

func (f *FieldWriter) Reset() {
	f.s = nil
	f.shouldMore = false
}

// Write "," (if needed) and new field.
func (f *FieldWriter) Write(k string) {
	if f.shouldMore {
		f.s.WriteMore()
		f.shouldMore = false
	}

	f.s.WriteObjectField(k)
	f.shouldMore = true
}
