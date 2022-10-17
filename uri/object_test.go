package uri

import "testing"

func BenchmarkEncodeObject(b *testing.B) {
	const kvSep, fieldSep = '=', ','
	fields := []Field{
		{Name: "key1", Value: "value1"},
		{Name: "key2", Value: "value2"},
		{Name: "key3", Value: "value3"},
		{Name: "key4", Value: "value4"},
		{Name: "key5", Value: "value5"},
		{Name: "key6", Value: "value6"},
	}
	b.ReportAllocs()
	b.ResetTimer()

	var sink string
	for i := 0; i < b.N; i++ {
		sink = encodeObject(kvSep, fieldSep, fields)
	}
	if sink == "" {
		b.Fatal(sink)
	}
}
