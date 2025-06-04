package ir

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExternalEncoding(t *testing.T) {
	tests := []struct {
		input   string
		want    ExternalType
		wantErr string
	}{
		{
			input: "time.Time",
			want: ExternalType{
				PackagePath: "time",
				TypeName:    "Time",
				Encode:      ExternalJSON | ExternalText | ExternalBinary,
				Decode:      ExternalJSON | ExternalText | ExternalBinary,
			},
		},
		{
			input: "*time.Time",
			want: ExternalType{
				PackagePath: "time",
				TypeName:    "Time",
				Encode:      ExternalJSON | ExternalText | ExternalBinary,
				Decode:      ExternalJSON | ExternalText | ExternalBinary,
				IsPointer:   true,
			},
		},
		{
			input: "github.com/ogen-go/ogen/_testdata/testtypes.NumberOgen",
			want: ExternalType{
				PackagePath: "github.com/ogen-go/ogen/_testdata/testtypes",
				TypeName:    "NumberOgen",
				Encode:      ExternalNative,
				Decode:      ExternalNative,
			},
		},
		{
			input: "*(github.com/ogen-go/ogen/_testdata/testtypes).NumberOgen",
			want: ExternalType{
				PackagePath: "github.com/ogen-go/ogen/_testdata/testtypes",
				TypeName:    "NumberOgen",
				Encode:      ExternalNative,
				Decode:      ExternalNative,
				IsPointer:   true,
			},
		},
		{
			input: "net.IPMask",
			want:  ExternalType{PackagePath: "net", TypeName: "IPMask"},
		},
		{
			input:   "foo/bar.Baz",
			wantErr: "type not found",
		},
		{
			input:   "*(foo/bar)Baz",
			wantErr: "expected '.' after ')'",
		},
		{
			input:   "*(foo/bar.Baz",
			wantErr: "unmatched '('",
		},
		{
			input:   "foo/bar",
			wantErr: "no '.' found in type path",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("[%s]", test.input), func(t *testing.T) {
			got, err := getExternalType(test.input)
			if test.wantErr != "" {
				require.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.want, got, "want: %+v, got: %+v", test.want, got)
		})
	}
}
