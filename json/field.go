package json

import (
	json "github.com/json-iterator/go"
)

// FieldWriter is helper for writing fields with ",".
type FieldWriter struct {
	shouldMore bool
}

// Write "," (if needed) and new field.
func (f *FieldWriter) Write(s *json.Stream, k string) {
	if f.shouldMore {
		s.WriteMore()
		f.shouldMore = false
	}

	s.WriteObjectField(k)
	f.shouldMore = true
}
