//go:build go1.18

package json

import json "github.com/json-iterator/go"

type ReadWrite interface {
	WriteJSON(s *json.Stream)
	ReadJSON(i *json.Iterator) bool
}

type Optional[T ReadWrite] struct {
	Value T
	Set   bool
}

func (o *Optional[T]) SetTo(v T) {
	o.Set = true
	o.Value = v
}

func (o Optional[T]) IsSet() bool { return o.Set }

func (o Optional[T]) WriteFieldJSON(k string, s *json.Stream) {
	if o.IsSet() {
		s.WriteObjectField(k)
		o.WriteJSON(s)
	}
}

func (o Optional[T]) WriteJSON(s *json.Stream) {
	o.Value.WriteJSON(s)
}

func (o Optional[T]) ReadJSON(i *json.Iterator) bool {
	panic("implement me")
}

