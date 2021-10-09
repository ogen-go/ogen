package json

import (
	"testing"

	json "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
)

func TestParseSlice(t *testing.T) {
	i := json.NewIterator(json.ConfigDefault)
	i.ResetBytes([]byte(`[1, 2, 3.14, "foo", "bar", 15]`))
	var result []interface{}
	i.ReadArrayCB(func(i *json.Iterator) bool {
		switch i.WhatIsNext() {
		case json.NumberValue:
			n := i.ReadNumber()
			if v, err := n.Int64(); err == nil {
				result = append(result, v)
				return true
			}
			v, err := n.Float64()
			if err != nil {
				i.ReportError("ParseNumber", err.Error())
				return false
			}
			result = append(result, v)
			return true
		default:
			result = append(result, i.Read())
		}
		return true
	})
	require.NoError(t, i.Error)
	for _, v := range result {
		t.Logf("%v %T", v, v)
	}
}
