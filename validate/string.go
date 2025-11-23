package validate

import (
	"fmt"
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

	// Numeric constraints for strings representing numbers
	MinNumeric    float64
	MinNumericSet bool
	MaxNumeric    float64
	MaxNumericSet bool
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

// SetMaximumNumeric sets maximum numeric value for numeric strings.
func (t *String) SetMaximumNumeric(v float64) {
	t.MaxNumericSet = true
	t.MaxNumeric = v
}

// SetMinimumNumeric sets minimum numeric value for numeric strings.
func (t *String) SetMinimumNumeric(v float64) {
	t.MinNumericSet = true
	t.MinNumeric = v
}

// Set reports whether any validations are set.
func (t String) Set() bool {
	return t.MaxLengthSet || t.MinLengthSet || t.Email || t.Regex != nil || t.Hostname || t.MinNumericSet || t.MaxNumericSet
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
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && (r < 'A' || r > 'Z') {
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
			return &NoRegexMatchError{
				Pattern: r,
			}
		}
	}
	// Validate numeric constraints on string values
	if t.MinNumericSet || t.MaxNumericSet {
		if err := t.validateNumeric(v); err != nil {
			return err
		}
	}
	return nil
}

func (t String) validateNumeric(v string) error {
	// Parse string as float64
	var val float64
	if _, err := fmt.Sscanf(v, "%f", &val); err != nil {
		return errors.Wrap(err, "parse as number")
	}

	if t.MinNumericSet && val < t.MinNumeric {
		return errors.Errorf("value %f less than minimum %f", val, t.MinNumeric)
	}
	if t.MaxNumericSet && val > t.MaxNumeric {
		return errors.Errorf("value %f greater than maximum %f", val, t.MaxNumeric)
	}
	return nil
}
