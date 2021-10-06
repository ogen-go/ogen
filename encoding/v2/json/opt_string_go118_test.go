//go:build go1.18

package json

import (
	"testing"

	json "github.com/json-iterator/go"
)

type String string

func (v String) WriteFieldJSON(k string, s *json.Stream) {
	s.WriteObjectField(k)
	v.WriteJSON(s)
}

func (v String) WriteJSON(s *json.Stream) {
	s.WriteString(string(v))
}

func (v *String) ReadJSON(i *json.Iterator) bool {
	panic("implement me")
}

func TestOptional(t *testing.T) {
	var v Optional[String]
	v.SetTo("Hello, world!")
}
