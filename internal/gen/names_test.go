package gen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNames(t *testing.T) {
	tests := []struct {
		Input   string
		AllowMP bool
		Expect  string
	}{
		{
			Input:  "user_id",
			Expect: "UserID",
		},
		{
			Input:   "foo+bar",
			AllowMP: true,
			Expect:  "FooPlusBar",
		},
		{
			Input:   "+1",
			AllowMP: true,
			Expect:  "Plus1",
		},
		{
			Input:   "foo+bar",
			AllowMP: false,
			Expect:  "FooBar",
		},
		{
			Input:  "userId",
			Expect: "UserId",
		},
	}

	for _, test := range tests {
		out := (&nameGen{
			src:     []rune(test.Input),
			allowMP: test.AllowMP,
		}).generate()
		require.Equal(t, test.Expect, out)
	}
}
