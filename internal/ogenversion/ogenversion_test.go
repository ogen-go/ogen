package ogenversion

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInfo_String(t *testing.T) {
	zone, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	date := time.Date(2001, 9, 11, 8, 46, 0, 0, zone)

	tests := []struct {
		info Info
		want string
	}{
		{
			Info{},
			`ogen version unknown`,
		},
		{
			Info{
				GoVersion: "go1.19.1",
			},
			`ogen version unknown (built with go1.19.1)`,
		},
		{
			Info{
				Version:   "v1.2.3",
				GoVersion: "go1.19.1",
			},
			`ogen version v1.2.3 (built with go1.19.1)`,
		},
		{
			Info{
				Version: "v1.2.3",
				Time:    date,
			},
			`ogen version v1.2.3 (built at Tue, 11 Sep 2001 08:46:00 EDT)`,
		},
		{
			Info{
				Version:   "v1.2.3",
				GoVersion: "go1.19.1",
				Commit:    "abcdef",
			},
			`ogen version v1.2.3-abcdef (built with go1.19.1)`,
		},
		{
			Info{
				Version:   "v1.2.3",
				GoVersion: "go1.19.1",
				Time:      date,
			},
			`ogen version v1.2.3 (built with go1.19.1 at Tue, 11 Sep 2001 08:46:00 EDT)`,
		},
		{
			Info{
				Version:   "v1.2.3",
				GoVersion: "go1.19.1",
				Commit:    "abcdef",
				Time:      date,
			},
			`ogen version v1.2.3-abcdef (built with go1.19.1 at Tue, 11 Sep 2001 08:46:00 EDT)`,
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			got := tt.info.String()
			// Cut off the GOOS and GOARCH, to make the test multi-platform.
			got = strings.TrimSuffix(got, " "+runtime.GOOS+"/"+runtime.GOARCH)
			a.Equal(tt.want, got)
		})
	}
}
