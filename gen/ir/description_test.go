package ir

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/internal/naming"
)

func Test_prettyDoc(t *testing.T) {
	tests := []struct {
		input  string
		notice string
		wantR  []string
	}{
		{
			input: ``,
			wantR: nil,
		},
		{
			input: `example`,
			wantR: []string{`Example.`},
		},
		{
			input:  `example`,
			notice: "Deprecated: for some reason.",
			wantR:  []string{`Example.`, ``, `Deprecated: for some reason.`},
		},
		{
			input: "The name of the fully qualified reference (ie: `refs/heads/master`). " +
				"If it doesn't start with 'refs' and have at least two slashes, it will be rejected.",
			wantR: []string{
				"The name of the fully qualified reference (ie: `refs/heads/master`). If it doesn't start with 'refs'",
				"and have at least two slashes, it will be rejected.",
			},
		},
		{
			// Markdown links are rendered as godoc references with the URL
			// collected into a link definition instead of being split mid-URL.
			input: "When this object was tagged. This is a timestamp in [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601) format: `YYYY-MM-DDTHH:MM:SSZ`.",
			wantR: []string{
				"When this object was tagged. This is a timestamp in [ISO 8601] format: `YYYY-MM-DDTHH:MM:SSZ`.",
				"",
				"[ISO 8601]: https://en.wikipedia.org/wiki/ISO_8601",
			},
		},
		{
			input: "Invite people to an organization by using their GitHub user ID or their email address. " +
				"In order to create invitations in an organization, the authenticated user must be an organization owner.\n\n" +
				"This endpoint triggers [notifications](https://docs.github.com/en/github/managing-subscriptions-and-notifications-on-github/about-notifications). " +
				"Creating content too quickly using this endpoint may result in secondary rate limiting.",
			wantR: []string{
				"Invite people to an organization by using their GitHub user ID or their email address. In order to",
				"create invitations in an organization, the authenticated user must be an organization owner.",
				"",
				"This endpoint triggers [notifications]. Creating content too quickly using this endpoint may result",
				"in secondary rate limiting.",
				"",
				"[notifications]: https://docs.github.com/en/github/managing-subscriptions-and-notifications-on-github/about-notifications",
			},
		},
		{
			// Headings, lists and code spans are rendered using godoc conventions.
			input: "# Heading\n\n" +
				"Some text with a [link](https://example.com).\n\n" +
				"- item one\n- item two with `code`\n\n" +
				"1. first\n2. second",
			wantR: []string{
				"# Heading",
				"",
				"Some text with a [link].",
				"",
				" - item one",
				" - item two with `code`",
				"",
				"1. first",
				"2. second",
				"",
				"[link]: https://example.com",
			},
		},
		{
			// Fenced code blocks are rendered as preformatted text.
			input: "Code example:\n\n```go\nfunc main() {}\n```",
			wantR: []string{
				"Code example:",
				"",
				"\tfunc main() {}",
			},
		},
		{
			// GFM tables are rendered as preformatted, aligned text.
			input: "| Name | Type |\n|------|------|\n| id   | int  |\n| name | str  |",
			wantR: []string{
				"\tName | Type",
				"\t-----+-----",
				"\tid   | int",
				"\tname | str",
			},
		},
		{
			input: strings.Repeat("a", lineLimit-4) + string(rune(12288)) + strings.Repeat("a", 10),
			wantR: []string{
				naming.Capitalize(strings.Repeat("a", lineLimit-4)),
				strings.Repeat("a", 10) + ".",
			},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			require.Equal(t, tt.wantR, prettyDoc(tt.input, tt.notice))
		})
	}
}
