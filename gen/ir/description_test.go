package ir

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/internal/naming"
)

func Test_prettyDoc(t *testing.T) {
	// Save the original value to restore after the test
	originalLimit := lineLimit
	defer func() {
		lineLimit = originalLimit
	}()

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
		{
			input: strings.Repeat("a", lineLimit-4) + string(rune(12288)) + strings.Repeat("a", 10),
			wantR: []string{
				naming.Capitalize(strings.Repeat("a", lineLimit-4)),
				strings.Repeat("a", 10) + ".",
			},
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			require.Equal(t, tt.wantR, prettyDoc(tt.input, tt.notice))
		})
	}
}

func TestSetLineLimit(t *testing.T) {
	// Save the original value to restore after the test
	originalLimit := lineLimit
	defer func() {
		lineLimit = originalLimit
	}()

	const longText = "This is a very long description that should be split into multiple lines depending on the configured line limit"

	tests := []struct {
		name     string
		limit    int
		expected int // expected number of lines (approximate)
	}{
		{
			name:     "default_limit",
			limit:    0, // Should use default 100
			expected: 2, // The long test will still need to be split in 2 with default limit
		},
		{
			name:     "short_limit",
			limit:    20,
			expected: 8, // Will split into multiple lines
		},
		{
			name:     "negative_limit_disables_wrapping",
			limit:    -1,
			expected: 1, // Should not wrap
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLineLimit(tt.limit)
			result := prettyDoc(longText, "")
			require.Len(t, result, tt.expected)
		})
	}
}

func TestCommentLineLimit(t *testing.T) {
	// Save the original value to restore after the test
	originalLimit := lineLimit
	defer func() {
		lineLimit = originalLimit
	}()

	// Reset to default to start
	SetLineLimit(0)
	require.Equal(t, 100, lineLimit, "Default line limit should be 100")

	// Test with a custom line limit
	SetLineLimit(50)
	require.Equal(t, 50, lineLimit, "Line limit should be updated to 50")

	// Test with a negative value (disabled wrapping)
	SetLineLimit(-1)
	require.Equal(t, -1, lineLimit, "Negative line limit should disable wrapping")

	// Test with zero (should set to default)
	SetLineLimit(0)
	require.Equal(t, 100, lineLimit, "Zero line limit should reset to default (100)")
}
