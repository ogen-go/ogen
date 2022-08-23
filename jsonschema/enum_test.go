package jsonschema

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestEnum(t *testing.T) {
	create := func() any {
		return &Enum{}
	}
	enum := func(s ...string) (r Enum) {
		for _, v := range s {
			r = append(r, json.RawMessage(v))
		}
		return r
	}

	tests := []struct {
		data    string
		value   Enum
		wantErr bool
	}{
		{`[0, 1]`, enum(`0`, `1`), false},
		{`[0, "string"]`, enum(`0`, `"string"`), false},
		{`[{}, []]`, enum(`{}`, `[]`), false},
		// Invalid YAML.
		{`"`, nil, true},
		{`[`, nil, true},
		// Invalid type.
		{`{}`, nil, true},
		{`"100"`, nil, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testCustomEncodings(create, tt.data, tt.wantErr))
	}
}
