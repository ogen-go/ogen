package main

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getRootURL(t *testing.T) {
	github := func(u string) FileMatch {
		return FileMatch{
			File: File{
				ExternalURLs: []ExternalURL{
					{ServiceKind: "GITHUB", URL: u},
				},
			},
		}
	}
	gitlab := func(u string) FileMatch {
		return FileMatch{
			File: File{
				ExternalURLs: []ExternalURL{
					{ServiceKind: "GITLAB", URL: u},
				},
			},
		}
	}
	mustURL := func(s string) *url.URL {
		u, err := url.Parse(s)
		require.NoError(t, err)
		return u
	}

	tests := []struct {
		m     FileMatch
		want  *url.URL
		want1 bool
	}{
		// GitHub links.
		{
			github("https://github.com/owner/repo/blob/abc-the-commit/dir1/dir2/file.yml"),
			mustURL("https://raw.githubusercontent.com/owner/repo/abc-the-commit/dir1/dir2/file.yml"),
			true,
		},
		{
			github("https://github.com/owner/repo/blob/abc-the-commit/file.yml"),
			mustURL("https://raw.githubusercontent.com/owner/repo/abc-the-commit/file.yml"),
			true,
		},
		// Bad GitHub links.
		{
			github("https://github.com/owner/blob/abc-the-commit/file.yml"),
			nil,
			false,
		},

		// GitLab links.
		{
			gitlab("https://gitlab.com/owner/repo/blob/abc-the-commit/file.yml"),
			mustURL("https://gitlab.com/owner/repo/raw/abc-the-commit/file.yml"),
			true,
		},
		{
			gitlab("https://gitlab.com/owner/repo/-/blob/abc-the-commit/file.yml"),
			mustURL("https://gitlab.com/owner/repo/raw/abc-the-commit/file.yml"),
			true,
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			a := require.New(t)

			got, ok := getRootURL(tt.m)
			if !tt.want1 {
				a.False(ok)
				return
			}
			a.True(ok)
			a.Equal(tt.want, got)
		})
	}
}
