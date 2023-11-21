package validate

import (
	"unicode"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/ogenregex"
)

// String validator.
type String struct {
	MinLength    int
	MinLengthSet bool
	MaxLength    int
	MaxLengthSet bool
	Email        bool
	Regex        ogenregex.Regexp
	Hostname     bool
}

// SetMaxLength sets maximum string length (in Unicode code points).
func (t *String) SetMaxLength(v int) {
	t.MaxLengthSet = true
	t.MaxLength = v
}

// SetMinLength sets minimum string length (in Unicode code points).
func (t *String) SetMinLength(v int) {
	t.MinLengthSet = true
	t.MinLength = v
}

// Set reports whether any validations are set.
func (t String) Set() bool {
	return t.MaxLengthSet || t.MinLengthSet || t.Email || t.Regex != nil || t.Hostname
}

func (t String) checkHostname(v string) error {
	if v == "" {
		return errors.New("blank")
	}
	if len([]rune(v)) >= 255 {
		return errors.New("too long")
	}
	for _, r := range v {
		if r == '.' {
			continue
		}
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r >= 'A' && r <= 'Z') {
			if unicode.IsSpace(r) {
				return errors.Errorf("space character (%U)", r)
			}
			if !unicode.IsPrint(r) {
				return errors.Errorf("not printable character (%U)", r)
			}
			return errors.Errorf("invalid character (%U)", r)
		}
	}
	return nil
}

func (t String) checkEmail(v string) error {
	// Pretty basic validation, but should work for most cases and is not
	// too strict to break things.
	//
	// Still better than obscure regex or std `mail.ParseAddress`.
	var (
		gotAt bool
		last  rune
	)
	for i, r := range v {
		if unicode.IsSpace(r) {
			return errors.Errorf("space character (%U)", r)
		}
		if !unicode.IsPrint(r) {
			return errors.Errorf("not printable character (%U)", r)
		}

		last = r
		if r != '@' {
			continue
		}
		if gotAt {
			return errors.New(`got @ multiple times`)
		}
		if i == 0 {
			return errors.New(`got @ at start`)
		}
		gotAt = true
	}
	if last == '@' {
		return errors.New("@ at end")
	}
	if !gotAt {
		return errors.New(`no @`)
	}
	return nil
}

// Validate returns error if v does not match validation rules.
func (t String) Validate(v string) error {
	if err := (Array{
		MinLength:    t.MinLength,
		MinLengthSet: t.MinLengthSet,
		MaxLength:    t.MaxLength,
		MaxLengthSet: t.MaxLengthSet,
	}).ValidateLength(len([]rune(v))); err != nil {
		return err
	}
	if t.Email {
		if err := t.checkEmail(v); err != nil {
			return err
		}
	}
	if t.Hostname {
		if err := t.checkHostname(v); err != nil {
			return err
		}
	}
	if r := t.Regex; r != nil {
		match, err := r.MatchString(v)
		if err != nil {
			return errors.Wrap(err, "execute regex")
		}
		if !match {
			return &NoRegexMatchError{}
		}
	}
	return nil
}
