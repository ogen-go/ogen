//go:build !go1.18

package json

import (
	json "github.com/json-iterator/go"
)

type World struct {
	Key   OptionalNullableString
	Value OptionalNullableString
}

func (w World) WriteFieldJSON(k string, s *json.Stream) {
	s.WriteObjectField(k)
}

func (w World) WriteJSON(s *json.Stream) {
	s.WriteObjectStart()
	s.WriteObjectEnd()
}

func (w World) ReadJSON(i *json.Iterator) bool {
	panic("implement me")
}
