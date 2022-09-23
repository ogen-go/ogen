package jsonschema

import (
	"fmt"
	"testing"

	yaml "github.com/go-faster/yamlx"
	"github.com/stretchr/testify/require"
)

func Test_convertYAMLtoRawJSON(t *testing.T) {
	mustNode := func(s string) *yaml.Node {
		var n yaml.Node
		if err := yaml.Unmarshal([]byte(s), &n); err != nil {
			t.Fatal(err)
			return nil
		}
		return &n
	}
	tests := []struct {
		node    *yaml.Node
		want    string
		wantErr bool
	}{
		{mustNode(`1`), `1`, false},
		{mustNode(`"foo"`), `"foo"`, false},
		{mustNode(`foo: bar`), `{"foo": "bar"}`, false},
		{mustNode(`foo: [1, 2, 3]`), `{"foo": [1, 2, 3]}`, false},
		{mustNode(`10: foo`), `{"10": "foo"}`, false},
		{&yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Tag: "!!int", Value: "10"},
				{Kind: yaml.ScalarNode, Tag: "!!str", Value: "foo"},
			},
		}, `{"10": "foo"}`, false},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			got, err := convertYAMLtoRawJSON(tt.node)
			if tt.wantErr {
				a.Error(err)
				return
			}
			a.NoError(err)
			a.JSONEq(tt.want, string(got))
		})
	}
}
