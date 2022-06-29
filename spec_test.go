package ogen

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPropertiesJSON(t *testing.T) {
	var props Properties
	data := `{"foo":{"type":"integer"},"bar":{"type":"string"}}`
	require.NoError(t, json.Unmarshal([]byte(data), &props))

	expect := Properties{
		{
			Name:   "foo",
			Schema: &Schema{Type: "integer"},
		},
		{
			Name:   "bar",
			Schema: &Schema{Type: "string"},
		},
	}
	require.Equal(t, expect, props)

	b, err := json.Marshal(props)
	require.NoError(t, err)
	require.Equal(t, data, string(b))
}
