package json

import (
	"fmt"
	"sort"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"
)

func TestEqual(t *testing.T) {
	type testCase struct {
		a, b    string
		want    bool
		wantErr bool
	}
	tests := map[string][]testCase{}

	for _, c := range []struct {
		a, b string
		want bool
	}{
		// Null.
		{`null`, `null`, true},
		// Bool.
		{`false`, `false`, true},
		{`true`, `true`, true},
		{`false`, `true`, false},
		// String.
		{`"foo"`, `"foo"`, true},
		{`"foo"`, `"foo" `, true},
		{`"foo\u000a"`, `"foo\n"`, true},
		{`"foo"`, `"foo\n"`, false},
		{`"foo"`, `"foo "`, false},
		// Number.
		{`0`, `0`, true},
		{`-0`, `-0`, true},
		{`-0`, `0`, true},
		{`1`, `1`, true},
		{`10`, `10`, true},
		{`0.0`, `0.0`, true},
		{`10`, `1e1`, true},
		{` 1000000000000000000000000000000`, `1000000000000000000000000000000`, true},
		{`1e1000`, `1.0e1000`, true},
		{`10`, `1.0e1`, true},
		{`0`, `1`, false},
		{`-1`, `1`, false},
		{`1e1`, `100`, false},
		// Array.
		{`[]`, `[]`, true},
		{`[]`, `[ ]`, true},
		{`[[]]`, `[[] ]`, true},
		{`["a", "b"]`, `["a", "b"]`, true},
		{`["a"]`, `[]`, false},
		{`[1,2,3]`, `[1,2]`, false},
		{`[[]]`, `[[1]]`, false},
		{`["b","a"]`, `["a","b"]`, false},
		// Object.
		{`{}`, `{}`, true},
		{`{}`, `{ }`, true},
		{`{"a":"b"}`, `{"a":"b"}`, true},
		{`{"a":"b","b":"a"}`, `{"b":"a", "a":"b"}`, true},
		{`{}`, `{"a":"b"}`, false},
		{`{"b":"a"}`, `{"a":"b"}`, false},
		{`{"a":10}`, `{"a":"b"}`, false},
		// Type comparison.
		{`{}`, `[]`, false},
		{`{}`, `0`, false},
		{`{}`, `null`, false},
		{`{}`, `false`, false},
		{`{}`, `""`, false},
	} {
		typ := jx.DecodeStr(c.a).Next().String()
		typb := jx.DecodeStr(c.b).Next().String()
		require.NotContains(t, []string{typ, typb}, jx.Invalid.String())
		if typb != typ {
			typ = "type"
		}
		tests[typ] = append(tests[typ], testCase{
			a:       c.a,
			b:       c.b,
			want:    c.want,
			wantErr: false,
		})
	}

	sortedIter := func(cb func(k string, tts []testCase)) {
		var keys []string
		for k := range tests {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			cb(k, tests[k])
		}
	}

	sortedIter(func(typ string, tts []testCase) {
		t.Run(typ, func(t *testing.T) {
			for i, tt := range tts {
				tt := tt
				t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
					a := require.New(t)
					check := func(x, y []byte) {
						got, err := Equal(x, y)
						if tt.wantErr {
							a.Error(err)
							a.False(got)
							return
						}
						a.NoError(err)
						a.Equal(tt.want, got, "%q == %q must be %v", x, y, tt.want)
					}

					x, y := []byte(tt.a), []byte(tt.b)
					check(x, y)
					check(y, x)
				})
			}
		})
	})
}

func BenchmarkEqual(b *testing.B) {
	benches := []struct {
		name string
		a, b string
		sink bool
	}{
		{`Zero`, `0`, `0`, true},
		{`SmallInt`, `10`, `21`, false},
		{`BigInt`, `1000000000000000000000000`, `21000000000000000000000000`, false},
		{`SmallFloat`, `15.20`, `69.12`, false},
		{`BigFloat`, `1e100`, `1e101`, false},
		{`PlainString`, `"foo10bar"`, `"foobar"`, false},
		{`EscapedString`, `"foo\nbar"`, `"foobar"`, false},
		{`Bool`, `true`, `true`, true},
		{`Array`, `[1,2,3,5,6,732,4312]`, `[1,2,3,5,6,732,4312]`, true},
		{`Object`, `{"b":"a","a":"b","foo":"bar"}'`, `{"foo":"bar","a":"b","b":"a"}`, true},
	}

	for _, bb := range benches {
		bb := bb
		b.Run(bb.name, func(b *testing.B) {
			x, y := []byte(bb.a), []byte(bb.b)
			var (
				sink bool
				err  error
			)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				sink, err = Equal(x, y)
			}

			if err != nil {
				b.Fatal(err)
			}
			if bb.sink != sink {
				b.Fatalf("%q == %q must be %v", x, y, bb.sink)
			}
		})
	}
}
