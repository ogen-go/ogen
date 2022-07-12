package gen

import (
	"go/token"
	"math"
	"strconv"
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

			name := setName(g.parts)

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

func setName(parts []string) string {
	name := strings.Join(parts, "")
	var intPrefix strings.Builder
	pos := 0
	for _, c := range []rune(name) {
		if unicode.IsDigit(c) {
			intPrefix.WriteRune(c)
			pos++
		} else {
			break
		}
	}
	var res strings.Builder
	res.Grow(intPrefix.Len() + len(name) - pos)
	convIntPrefix, err := strconv.Atoi(intPrefix.String())
	if err == nil {
		res.WriteString(convertIntegerToWord(convIntPrefix))
	}
	res.WriteString(name[pos:])
	return res.String()
}

func convertIntegerToWord(integer int) string {
	if integer == 0 {
		return "Zero"
	}
	var res strings.Builder
	if integer < 0 {
		res.WriteString("Minus")
		integer = int(math.Abs(float64(integer)))
	}

	var triplets []int

	for integer > 0 {
		triplets = append(triplets, integer%1000)
		integer /= 1000
	}

	megaWords := []string{"", "Thousand", "Million", "Billion", "Trillion", "Quadrillion", "Quintillion"}
	unitWords := []string{"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine"}
	tenWords := []string{"", "Ten", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}
	teenWords := []string{"Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen", "Seventeen", "Eighteen", "Nineteen"}

	for i := len(triplets) - 1; i >= 0; i-- {
		triplet := triplets[i]

		if triplet == 0 {
			continue
		}

		hundreds := triplet / 100 % 10
		tens := triplet / 10 % 10
		units := triplet % 10

		if hundreds > 0 {
			res.WriteString(unitWords[hundreds] + "Hundred")
		}

		switch tens {
		case 0:
			res.WriteString(unitWords[units])
		case 1:
			res.WriteString(teenWords[units])
		default:
			if units > 0 {
				res.WriteString(tenWords[tens] + unitWords[units])
			} else {
				res.WriteString(tenWords[tens])
			}
		}

		if mega := megaWords[i]; mega != "" {
			res.WriteString(megaWords[i])
		}
	}

	return res.String()
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
	r = unicode.ToLower(r)
	return (r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9')
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
