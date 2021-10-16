package json

import (
	json "github.com/json-iterator/go"
)

type FieldWriter struct {
	shouldMore bool
}

func (f *FieldWriter) Write(s *json.Stream, k string) {
	if f.shouldMore {
		s.WriteMore()
		f.shouldMore = false
	}

	s.WriteObjectField(k)
	f.shouldMore = true
}
