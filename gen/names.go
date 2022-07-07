package gen

import (
	"go/token"
	"strings"
	"unicode"

	"github.com/go-faster/errors"
)

type nameGen struct {
	parts []string
	src   []rune
	pos   int

	allowSpecial bool // special characters like +, -, /
}

func (g *nameGen) next() (rune, bool) {
	if len(g.src) == g.pos {
		return rune(0), false
	}

	defer func() { g.pos++ }()
	return g.src[g.pos], true
}

var namedChar = map[rune][]rune{
	'+': []rune("Plus"),
	'-': []rune("Minus"),
	'/': []rune("Slash"),
	'<': []rune("Less"),
	'>': []rune("Greater"),
	'=': []rune("Eq"),
	'.': []rune("Dot"),
}

func (g *nameGen) generate() (string, error) {
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

			name := strings.Join(g.parts, "")
			// FIXME(tdakkota): choose prefix according to context
			if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
				name = "R" + name
			}
			if !token.IsIdentifier(name) {
				return "", errors.Wrapf(&ErrNotImplemented{Name: "crypticName"},
					"can't generate valid name: %+v", g.parts)
			}
			return name, nil
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
		if g.allowSpecial {
			if p, ok := namedChar[r]; ok {
				pushPart()
				part = p
			}
		}

		pushPart()
	}
}

func (g *nameGen) clean() string {
	var (
		part     []rune
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
			part = append(part, r)
			continue
		}

		if g.allowSpecial {
			if p, ok := namedChar[r]; ok {
				pushPart()
				part = p
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

		"PNG", "JPG", "GIF", "MP4", "WEBP",
	}

	for _, rule := range rules {
		if strings.EqualFold(part, rule) {
			return rule
		}
	}

	return part
}

func cleanSpecial(strs ...string) string {
	return (&nameGen{
		src:          []rune(strings.Join(strs, " ")),
		allowSpecial: true,
	}).clean()
}

func pascal(strs ...string) (string, error) {
	return (&nameGen{
		src: []rune(strings.Join(strs, " ")),
	}).generate()
}

func pascalSpecial(strs ...string) (string, error) {
	return (&nameGen{
		src:          []rune(strings.Join(strs, " ")),
		allowSpecial: true,
	}).generate()
}

func pascalNonEmpty(strs ...string) (string, error) {
	r, err := pascal(strs...)
	if err == nil && r != "" {
		return r, nil
	}

	r, err = pascalSpecial(strs...)
	if err != nil {
		return "", err
	}
	if r != "" {
		return r, nil
	}
	return "", errors.Wrapf(&ErrNotImplemented{Name: "crypticName"}, "can't generate name for %+v", strs)
}

func camel(s ...string) (string, error) {
	r, err := pascal(s...)
	if err != nil {
		return "", err
	}
	return firstLower(r), nil
}

func camelSpecial(s ...string) (string, error) {
	r, err := pascalSpecial(s...)
	if err != nil {
		return "", err
	}
	return firstLower(r), nil
}

// firstLower returns s with first rune mapped to lower case.
func firstLower(s string) string {
	var out []rune
	for i, c := range s {
		if i == 0 {
			c = unicode.ToLower(c)
		}
		out = append(out, c)
	}
	return string(out)
}
