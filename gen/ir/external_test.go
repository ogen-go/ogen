package ir

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExternalEncoding(t *testing.T) {
	tests := []struct {
		pkg, typ   string
		wantEncode ExternalEncoding
		wantDecode ExternalEncoding
		wantErr    string
	}{
		{"time", "Time", ExternalJSON | ExternalText | ExternalBinary, ExternalJSON | ExternalText | ExternalBinary, ""},
		{"github.com/ogen-go/ogen/_testdata/testtypes", "NumberOgen", ExternalNative, ExternalNative, ""},
		{"net", "IPMask", 0, 0, ""},
		{"foo/bar", "Baz", -1, -1, "type not found: foo/bar.Baz"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("[%s.%s]", test.pkg, test.typ), func(t *testing.T) {
			enc, dec, err := getExternalEncoding(test.pkg, test.typ)
			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.wantEncode, enc, "want encoding %s, got %s", test.wantEncode, enc)
			assert.Equal(t, test.wantDecode, dec, "want decoding %s, got %s", test.wantDecode, dec)
		})
	}
}
