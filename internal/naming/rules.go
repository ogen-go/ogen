package naming

import (
	"maps"
	"strings"
)

var (
	rules = [...]string{
		"ACL", "API", "ASCII", "AWS", "CPU", "CSS", "DNS", "EOF", "GB", "GUID",
		"HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "KB", "LHS", "MAC", "MB",
		"QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SQL", "SSH", "SSO", "TLS",
		"TTL", "UI", "UID", "URI", "URL", "UTF8", "UUID", "VM", "XML", "XMPP",
		"XSRF", "XSS", "SMS", "CDN", "TCP", "UDP", "DC", "PFS", "P2P",
		"SHA256", "SHA1", "MD5", "SRP", "2FA", "OAuth", "OAuth2",

		"PNG", "JPG", "GIF", "MP4", "WEBP",
	}
	// defaultRuleset is the package-level ruleset used by [Rule].
	defaultRuleset = NewRuleset(rules[:]...)
)

// Ruleset maps lowered word parts to their canonical initialism form
// (e.g. "id" -> "ID"). Matching is case-insensitive.
//
// Use [DefaultRuleset] to get a copy of ogen's built-in initialisms, or
// [NewRuleset] to build a custom set from scratch.
type Ruleset struct {
	// m maps lowered rules to their canonical form.
	//
	// NOTE: we're using a map instead of a linear/binary search because
	// lowered string allocation is much cheaper than string comparison.
	// Also, ToLower doesn't allocate if the string is already in lower case.
	m map[string]string
}

// NewRuleset builds a Ruleset from the given canonical initialisms.
// Each initialism is matched case-insensitively against word parts.
func NewRuleset(initialisms ...string) *Ruleset {
	r := &Ruleset{m: make(map[string]string, len(initialisms))}
	for _, v := range initialisms {
		r.Add(v)
	}
	return r
}

// DefaultRuleset returns a fresh Ruleset containing ogen's built-in
// initialisms. The returned Ruleset is safe to mutate via [Ruleset.Add].
func DefaultRuleset() *Ruleset {
	return NewRuleset(rules[:]...)
}

// Add registers an initialism in the ruleset, overriding any existing rule
// that matches case-insensitively. Empty strings are ignored.
func (r *Ruleset) Add(initialism string) {
	if initialism == "" {
		return
	}
	r.m[strings.ToLower(initialism)] = initialism
}

// Merge copies all rules from other into r, overriding rules that match
// case-insensitively.
func (r *Ruleset) Merge(other *Ruleset) {
	maps.Copy(r.m, other.m)
}

// Rule returns the canonical initialism for the given part, if any.
// Otherwise, it returns ("", false).
func (r *Ruleset) Rule(part string) (string, bool) {
	v, ok := r.m[strings.ToLower(part)]
	return v, ok
}

// Rule returns the rule for the given part using the default ruleset, if any.
// Otherwise, it returns ("", false).
func Rule(part string) (string, bool) {
	return defaultRuleset.Rule(part)
}
