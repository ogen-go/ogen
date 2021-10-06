//go:build go1.18

package json

import json "github.com/json-iterator/go"

type ReadWrite interface {
	WriteJSON(s *json.Stream)
	ReadJSON(i *json.Iterator) bool
}

type String string

func (v String) WriteFieldJSON(k string, s *json.Stream) {
	s.WriteObjectField(k)
	v.WriteJSON(s)
}

func (v String) WriteJSON(s *json.Stream) {
	s.WriteString(string(v))
}

func (v *String) ReadJSON(i *json.Iterator) bool {
	s := i.ReadString()
	*v = String(s)

	return true
}

func (v *String) Set(s String) {
	*v = s
}

type OptionalValue interface {
	WriteJSON(s *json.Stream)
}

type Reader[T OptionalValue] interface {
	ReadJSON(i *json.Iterator) bool
	Set(T)
	*T
}

type Optional[T OptionalValue, S Reader[T]] struct {
	Value T
	Set   bool
}

func (o *Optional[T, S]) SetTo(v T) {
	o.Set = true
	settableValue := S(&o.Value)
	settableValue.Set(v)
}

func (o Optional[T, S]) IsSet() bool { return o.Set }

func (o Optional[T, S]) WriteFieldJSON(k string, s *json.Stream) {
	if o.IsSet() {
		s.WriteObjectField(k)
		o.WriteJSON(s)
	}
}

func (o Optional[T, S]) WriteJSON(s *json.Stream) {
	o.Value.WriteJSON(s)
}

func (o *Optional[T, S]) ReadJSON(i *json.Iterator) bool {
	settableValue := S(&o.Value)
	return settableValue.ReadJSON(i)
}

