package gen

import (
	"regexp"

	"github.com/ogen-go/ogen/gen/ir"
	"github.com/ogen-go/ogen/internal/xmaps"
	"github.com/ogen-go/ogen/ogenregex"
)

type TemplateConfig struct {
	Package       string
	Operations    []*ir.Operation
	Webhooks      []*ir.Operation
	Types         map[string]*ir.Type
	Interfaces    map[string]*ir.Type
	Error         *ir.Response
	ErrorType     *ir.Type
	Servers       ir.Servers
	Securities    map[string]*ir.Security
	Router        Router
	WebhookRouter WebhookRouter
	ClientEnabled bool
	ServerEnabled bool

	skipTestRegex *regexp.Regexp
}

// ErrorGoType returns Go type of error.
func (t TemplateConfig) ErrorGoType() string {
	typ := t.ErrorType
	if typ.DoPassByPointer() {
		return "*" + typ.Go()
	}
	return typ.Go()
}

// SkipTest returns true, if test should be skipped.
func (t TemplateConfig) SkipTest(typ *ir.Type) bool {
	return t.skipTestRegex != nil && t.skipTestRegex.MatchString(typ.Name)
}

func (t TemplateConfig) collectStrings(cb func(typ *ir.Type) []string) []string {
	var (
		add  func(typ *ir.Type)
		m    = map[string]struct{}{}
		seen = map[*ir.Type]struct{}{}
	)
	add = func(typ *ir.Type) {
		_, skip := seen[typ]
		if typ == nil || skip {
			return
		}
		seen[typ] = struct{}{}
		for _, got := range cb(typ) {
			m[got] = struct{}{}
		}

		for _, f := range typ.Fields {
			add(f.Type)
		}
		for _, f := range typ.SumOf {
			add(f)
		}
		add(typ.AliasTo)
		add(typ.PointerTo)
		add(typ.GenericOf)
		add(typ.Item)
	}

	for _, typ := range t.Types {
		add(typ)
	}
	for _, typ := range t.Interfaces {
		add(typ)
	}
	if t.Error != nil {
		add(t.Error.NoContent)
		for _, media := range t.Error.Contents {
			add(media.Type)
		}
	}
	add(t.ErrorType)

	_ = walkOpTypes(t.Operations, func(t *ir.Type) error {
		add(t)
		return nil
	})
	_ = walkOpTypes(t.Webhooks, func(t *ir.Type) error {
		add(t)
		return nil
	})

	return xmaps.SortedKeys(m)
}

// RegexStrings returns slice of all unique regex validators.
func (t TemplateConfig) RegexStrings() []string {
	return t.collectStrings(func(typ *ir.Type) (r []string) {
		for _, exp := range []ogenregex.Regexp{
			typ.Validators.String.Regex,
			typ.MapPattern,
		} {
			if exp == nil {
				continue
			}
			r = append(r, exp.String())
		}
		return r
	})
}

// RatStrings returns slice of all unique big.Rat (multipleOf validation).
func (t TemplateConfig) RatStrings() []string {
	return t.collectStrings(func(typ *ir.Type) []string {
		if r := typ.Validators.Float.MultipleOf; r != nil {
			// `RatString` return a string with integer value if denominator is 1.
			//
			// That makes string representation of `big.Rat` shorter and simpler.
			// Also, it is better for executable size.
			return []string{r.RatString()}
		}
		return nil
	})
}
