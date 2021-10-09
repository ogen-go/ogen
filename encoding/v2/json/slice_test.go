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
		result = append(result, i.Read())
		return true
	})
	require.NoError(t, i.Error)
	t.Log(result)
}
