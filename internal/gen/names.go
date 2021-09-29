package gen

import (
	"strings"
	"unicode"
)

func splitWords(s string) []string {
	allowed := func(r rune) bool {
		const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
		for _, c := range alphabet {
			if c == unicode.ToLower(r) {
				return true
			}
		}
		return false
	}

	var (
		words []string
		word  []rune
		push  = func() {
			if len(word) > 0 {
				words = append(words, string(word))
				word = nil
			}
		}
	)

	for _, r := range s {
		if !allowed(r) {
			push()
			continue
		}

		if unicode.IsUpper(r) {
			push()
		}

		word = append(word, r)
	}

	push()
	return words
}

func checkRules(word string) string {
	rules := []string{
		"ACL", "API", "ASCII", "AWS", "CPU", "CSS", "DNS", "EOF", "GB", "GUID",
		"HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "KB", "LHS", "MAC", "MB",
		"QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SQL", "SSH", "SSO", "TCP",
		"TLS", "TTL", "UDP", "UI", "UID", "URI", "URL", "UTF8", "UUID", "VM",
		"XML", "XMPP", "XSRF", "XSS", "SMS", "CDN", "TCP", "UDP", "DC", "PFS",
		"P2P", "SHA256", "SHA1", "MD5", "SRP", "2FA", "OAuth", "OAuth2",
	}

	for _, rule := range rules {
		if strings.EqualFold(word, rule) {
			return rule
		}
	}

	return word
}

func pascal(strs ...string) string {
	split := func(strs []string) (out []string) {
		for _, s := range strs {
			out = append(out, splitWords(s)...)
		}
		return
	}

	var out []string
	for _, s := range split(strs) {
		s = checkRules(s)
		rs := []rune(s)
		rs[0] = unicode.ToUpper(rs[0])
		out = append(out, string(rs))
	}
	return strings.Join(out, "")
}

func camel(s string) string {
	rs := []rune(pascal(s))
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}
