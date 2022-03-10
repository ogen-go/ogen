package gen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/jsonschema"
)

func Test_filterMostSpecific(t *testing.T) {
	sch := new(jsonschema.Schema)
	list := func(input ...string) (r map[string]*jsonschema.Schema) {
		r = map[string]*jsonschema.Schema{}
		for _, value := range input {
			r[value] = sch
		}
		return r
	}

	tests := []struct {
		contents map[string]*jsonschema.Schema
		expected map[string]*jsonschema.Schema
		wantErr  bool
	}{
		{list("text/html", "text/*"), list("text/html"), false},
		{list("text/*"), list("text/*"), false},
		{list("*/*"), list("*/*"), false},
		{list("text/html", "*/*"), list("text/html"), false},
		{list("text/*", "*/*"), list("text/*"), false},
		{list("text/html", "text/*", "*/*", "*"), list("text/html"), false},
		{list("*/json", "application/json"), list("application/json"), false},
		{list("*/json", "application/json", "text/json"), list("application/json", "text/json"), false},
		{list("*/json", "*application/json"), list("*application/json"), false},
		{list("application/*", "text/html"), list("application/*", "text/html"), false},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			contents := map[string]*jsonschema.Schema{}
			for k, v := range tt.contents {
				contents[k] = v
			}
			err := filterMostSpecific(contents)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, contents)
			}
		})
	}
}
