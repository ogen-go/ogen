package jsonschema

import (
	"testing"
)

func TestConst(t *testing.T) {
	create := func() any {
		return &Const{}
	}

	tests := []struct {
		name    string
		data    string
		value   Const
		wantErr bool
	}{
		{"number", `100`, Const(`100`), false},
		{"string", `"string"`, Const(`"string"`), false},
		{"object", `{"key": "value"}`, Const(`{"key": "value"}`), false},
		{"array", `[1, 2, 3]`, Const(`[1, 2, 3]`), false},
		// Invalid YAML.
		{"invalid_yaml_1", `"`, nil, true},
		{"invalid_yaml_2", `[`, nil, true},
		// Invalid type.
		{"invalid_type_1", `{}`, nil, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, testCustomEncodings(create, tt.data, tt.wantErr))
	}
}
