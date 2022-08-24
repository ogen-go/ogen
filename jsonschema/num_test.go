package jsonschema

import (
	"fmt"
	"testing"
)

func TestNum(t *testing.T) {
	create := func() any {
		return &Num{}
	}

	tests := []struct {
		data    string
		value   Num
		wantErr bool
	}{
		{`0`, Num(`0`), false},
		{`-1`, Num(`-1`), false},
		{`1.1`, Num(`1.1`), false},
		{`0.01e1`, Num(`0.1`), false},
		// Invalid YAML.
		{`"`, nil, true},
		{`0ee1`, nil, true},
		// Invalid type.
		{`{}`, nil, true},
		{`"100"`, nil, true},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), testCustomEncodings(create, tt.data, tt.wantErr))
	}
}
