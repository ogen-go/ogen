package ir

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_prettyDoc(t *testing.T) {
	tests := []struct {
		input string
		wantR []string
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
			input: "The name of the fully qualified reference (ie: `refs/heads/master`). " +
				"If it doesn't start with 'refs' and have at least two slashes, it will be rejected.",
			wantR: []string{
				"The name of the fully qualified reference (ie: `refs/heads/master`). If it doesn't start with",
				"'refs' and have at least two slashes, it will be rejected.",
			},
		},
		{
			input: "When this object was tagged. This is a timestamp in [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601) format: `YYYY-MM-DDTHH:MM:SSZ`.",
			wantR: []string{
				"When this object was tagged. This is a timestamp in [ISO 8601](https://en.wikipedia.",
				"org/wiki/ISO_8601) format: `YYYY-MM-DDTHH:MM:SSZ`.",
			},
		},
		{
			input: "Invite people to an organization by using their GitHub user ID or their email address. " +
				"In order to create invitations in an organization, the authenticated user must be an organization owner.\n\n" +
				"This endpoint triggers [notifications](https://docs.github.com/en/github/managing-subscriptions-and-notifications-on-github/about-notifications). " +
				"Creating content too quickly using this endpoint may result in secondary rate limiting." +
				" See \"[Secondary rate limits](https://docs.github.com/rest/overview/resources-in-the-rest-api#secondary-rate-limits)\" " +
				"and \"[Dealing with secondary rate limits](https://docs.github.com/rest/guides/best-practices-for-integrators#dealing-with-secondary-rate-limits)\"" +
				"for details.",
			wantR: []string{
				"Invite people to an organization by using their GitHub user ID or their email address. In order to",
				"create invitations in an organization, the authenticated user must be an organization owner.",
				"This endpoint triggers [notifications](https://docs.github.",
				"com/en/github/managing-subscriptions-and-notifications-on-github/about-notifications). Creating",
				"content too quickly using this endpoint may result in secondary rate limiting. See \"[Secondary",
				"rate limits](https://docs.github.", "com/rest/overview/resources-in-the-rest-api#secondary-rate-limits)\" and \"[Dealing with secondary",
				"rate limits](https://docs.github.", "com/rest/guides/best-practices-for-integrators#dealing-with-secondary-rate-limits)\"for details.",
			},
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			require.Equal(t, tt.wantR, prettyDoc(tt.input))
		})
	}
}
