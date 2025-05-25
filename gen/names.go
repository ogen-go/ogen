package gen

import (
	"fmt"
	"go/token"
	"iter"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

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

// valueMappingNameGen creates a name generator for either an enum or discriminator mapping
func valueMappingNameGen(
	mapType, name string,
	values iter.Seq[any],
	valuesLen int,
	allowSpecial bool,
) (func(v any, idx int) (string, error), error) {
	type namingStrategy int
	const (
		pascalName namingStrategy = iota
		pascalSpecialName
		cleanSuffix
		indexSuffix
		_lastStrategy
	)

	vstrCache := make(map[int]string, valuesLen)
	nameGen := func(s namingStrategy, v any, idx int) (string, error) {
		vstr, ok := vstrCache[idx]
		if !ok {
			vstr = fmt.Sprintf("%v", v)
			if vstr == "" {
				vstr = "Empty"
			}
			vstrCache[idx] = vstr
		}
		switch s {
		case pascalName:
			return pascal(name, vstr)
		case pascalSpecialName:
			return pascalSpecial(name, vstr)
		case cleanSuffix:
			return name + "_" + cleanSpecial(vstr), nil
		case indexSuffix:
			return name + "_" + strconv.Itoa(idx), nil
		default:
			panic(unreachable(s))
		}
	}

	isException := func(start namingStrategy) bool {
		if start == pascalName {
			// This code is called when vstrCache is fully populated, so it's ok.
			for _, v := range vstrCache {
				if v == "" {
					continue
				}

				// Do not use pascal strategy for enum values starting with special characters.
				//
				// This rule is created to be able to distinguish
				// between negative and positive numbers in this case:
				//
				// enum:
				//   - '1'
				//   - '-2'
				//   - '3'
				//   - '-4'
				firstRune, _ := utf8.DecodeRuneInString(v)
				if firstRune == utf8.RuneError {
					panic(fmt.Sprintf("invalid %s variant for %s: %q", mapType, name, v))
				}

				_, isFirstCharSpecial := namedChar[firstRune]
				if isFirstCharSpecial {
					return true
				}
			}
		}

		return false
	}

nextStrategy:
	for strategy := pascalName; strategy < _lastStrategy; strategy++ {
		if !allowSpecial && strategy == pascalSpecialName {
			continue nextStrategy
		}

		// Treat enum type name as duplicate to prevent collisions.
		names := map[string]struct{}{
			name: {},
		}
		idx := -1
		for v := range values {
			idx++
			k, err := nameGen(strategy, v, idx)
			if err != nil {
				continue nextStrategy
			}
			if _, ok := names[k]; ok {
				continue nextStrategy
			}
			names[k] = struct{}{}
		}
		if isException(strategy) {
			continue nextStrategy
		}
		return func(v any, idx int) (string, error) {
			return nameGen(strategy, v, idx)
		}, nil
	}
	return nil, errors.Errorf("unable to generate %s variant names for %q", mapType, name)
}

// enumVariantNameGen creates a name generator for enum values.
func enumVariantNameGen(name string, values []any) (func(v any, idx int) (string, error), error) {
	return valueMappingNameGen("enum", name, slices.Values(values), len(values), true)
}

// discriminatorMappingNameGen creates a name generator for discriminator mapping keys.
func discriminatorMappingNameGen(name string, keys []string) (func(v any, idx int) (string, error), error) {
	if len(keys) == 0 {
		return nil, errors.New("empty discriminator keys")
	}

	// Create sequence that yields the mapping keys
	seq := func(yield func(any) bool) {
		for _, key := range keys {
			// ensure user_id => UserId and not UserID
			key := strings.ReplaceAll(key, "_", "+")
			if !yield(key) {
				return
			}
		}
	}

	return valueMappingNameGen("discriminator", name, seq, len(keys), false)
}
