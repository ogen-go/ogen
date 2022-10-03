package naming

import (
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
	// rulesMap is a map of lowered rules to their canonical form.
	//
	// NOTE: we're using a map instead of a linear/binary search because
	// lowered string allocation is much cheaper than string comparison.
	// Also, ToLower doesn't allocate if the string is already in lower case.
	rulesMap = func() (r map[string]string) {
		r = make(map[string]string)
		for _, v := range rules {
			r[strings.ToLower(v)] = v
		}
		return r
	}()
)

// Rule returns the rule for the given part, if any.
// Otherwise, it returns ("", false).
func Rule(part string) (string, bool) {
	v, ok := rulesMap[strings.ToLower(part)]
	return v, ok
}
