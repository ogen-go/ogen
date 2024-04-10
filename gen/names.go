package gen

import (
	"go/token"
	"net/url"
	"path"
	"strings"
	"unicode"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/internal/naming"
	"github.com/ogen-go/ogen/jsonschema"
)

func cleanRef(r jsonschema.Ref) string {
	ref := r.String()

	before, result, _ := strings.Cut(ref, "#")
	if result == "" {
		result = ref
		// Cuts file name.
		//
		// https://example.com/foo/bar.json -> bar
		// foo/bar.json -> bar
		if u, err := url.Parse(before); before != "" && err == nil {
			_, result = path.Split(u.Path)
			result = strings.TrimSuffix(result, path.Ext(result))
		}
	}

	idx := strings.LastIndexByte(result, '/')
	if idx < 0 {
		return result
	}
	if cut := result[idx+1:]; cut != "" {
		result = cut
	}
	return result
}

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
				return "", errors.Errorf("can't generate valid name: %+v", g.parts)
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
	r = unicode.ToLower(r)
	return (r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9')
}

func (g *nameGen) checkPart(part string) string {
	if rule, ok := naming.Rule(part); ok {
		return rule
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
	return "", errors.Errorf("can't generate name for %+v", strs)
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
