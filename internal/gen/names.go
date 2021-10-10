package gen

import (
	"strings"
	"unicode"
)

type nameGen struct {
	parts []string
	src   []rune
	pos   int

	allowMP bool
}

func (g *nameGen) next() (rune, bool) {
	if len(g.src) == g.pos {
		return rune(0), false
	}

	defer func() { g.pos++ }()
	return g.src[g.pos], true
}

func (g *nameGen) generate() string {
	var (
		part     []rune
		upper    = true
		pushPart = func() {
			g.parts = append(g.parts, g.checkPart(string(part)))
			part = nil
		}
	)
	for {
		r, ok := g.next()
		if !ok {
			pushPart()
			return strings.Join(g.parts, "")
		}

		if g.isAllowed(r) {
			if upper {
				r = unicode.ToUpper(r)
				upper = false
			}

			part = append(part, r)
			continue
		}

		upper = true
		if g.allowMP {
			switch r {
			case '+':
				pushPart()
				part = []rune("Plus")
			case '-':
				pushPart()
				part = []rune("Minus")
			case '/':
				pushPart()
				part = []rune("Slash")
			}
		}

		pushPart()
	}
}

func (g *nameGen) isAllowed(r rune) bool {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	for _, c := range alphabet {
		if c == unicode.ToLower(r) {
			return true
		}
	}
	return false
}

func (g *nameGen) checkPart(part string) string {
	rules := []string{
		"ACL", "API", "ASCII", "AWS", "CPU", "CSS", "DNS", "EOF", "GB", "GUID",
		"HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "KB", "LHS", "MAC", "MB",
		"QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SQL", "SSH", "SSO", "TCP",
		"TLS", "TTL", "UDP", "UI", "UID", "URI", "URL", "UTF8", "UUID", "VM",
		"XML", "XMPP", "XSRF", "XSS", "SMS", "CDN", "TCP", "UDP", "DC", "PFS",
		"P2P", "SHA256", "SHA1", "MD5", "SRP", "2FA", "OAuth", "OAuth2",
	}

	for _, rule := range rules {
		if strings.EqualFold(part, rule) {
			return rule
		}
	}

	return part
}

func pascal(strs ...string) string {
	return (&nameGen{
		src: []rune(strings.Join(strs, " ")),
	}).generate()
}

func pascalMP(strs ...string) string {
	return (&nameGen{
		src:     []rune(strings.Join(strs, " ")),
		allowMP: true,
	}).generate()
}

func camel(s string) string {
	rs := []rune(pascal(s))
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}
