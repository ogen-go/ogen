//go:build go1.18

package json

import (
	"bytes"
	"testing"

	json "github.com/json-iterator/go"
)


func TestOptional(t *testing.T) {
	var v Optional[String, *String]
	v.SetTo("Hello, world!")
}

func BenchmarkOptional_ReadJSON(b *testing.B) {
	var v Optional[String, *String]
	buf := new(bytes.Buffer)
	buf.WriteString(`"foo"`)
	iter := json.NewIterator(json.ConfigFastest)

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		iter.ResetBytes(buf.Bytes())
		v.ReadJSON(iter)
		if err := iter.Error; err != nil {
			b.Error(err)
		}
	}
}
