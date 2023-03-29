package jsonschema

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-faster/yaml"
	"github.com/stretchr/testify/require"
)

func TestExtensions(t *testing.T) {
	create := func() any {
		return &Extensions{}
	}

	tests := []struct {
		data    string
		value   Extensions
		wantErr bool
	}{
		{`{"x-foo":"bar"}`, Extensions{
			"x-foo": yaml.Node{Kind: yaml.ScalarNode, Value: "bar"},
		}, false},
		// Invalid YAML.
		{`{`, Extensions{}, true},
		{`{]`, Extensions{}, true},
		// Invalid type.
		{`[]`, Extensions{}, true},
		{`0`, Extensions{}, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testCustomEncodings(create, tt.data, tt.wantErr))
	}
}

func testExtensionsMarshal(
	t *testing.T,
	marshal func(any) ([]byte, error),
	compare func(*require.Assertions, string, string, ...interface{}),
) {
	a := require.New(t)
	e := Extensions{
		"x-foo": {Kind: yaml.ScalarNode, Value: "baz"},
		"foo":   {Kind: yaml.ScalarNode, Value: "bar"},
	}
	data, err := marshal(e)
	a.NoError(err)
	compare(a, `{"x-foo":"baz"}`, string(data))
}

func TestExtensions_MarshalYAML(t *testing.T) {
	testExtensionsMarshal(t, yaml.Marshal, (*require.Assertions).YAMLEq)
}

func TestExtensions_MarshalJSON(t *testing.T) {
	testExtensionsMarshal(t, json.Marshal, (*require.Assertions).JSONEq)
}
