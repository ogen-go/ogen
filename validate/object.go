package validate

import "github.com/go-faster/errors"

// Object validates map length.
type Object struct {
	MinProperties    int
	MinPropertiesSet bool
	MaxProperties    int
	MaxPropertiesSet bool
	MinLength        int
	MinLengthSet     bool
	MaxLength        int
	MaxLengthSet     bool
}

// SetMinProperties sets MinProperties validation.
func (t *Object) SetMinProperties(v int) {
	t.MinPropertiesSet = true
	t.MinProperties = v
}

// SetMaxProperties sets MaxProperties validation.
func (t *Object) SetMaxProperties(v int) {
	t.MaxPropertiesSet = true
	t.MaxProperties = v
}

// SetMinLength sets MinLength validation.
func (t *Object) SetMinLength(v int) {
	t.MinLengthSet = true
	t.MinLength = v
}

// SetMaxLength sets MaxLength validation.
func (t *Object) SetMaxLength(v int) {
	t.MaxLengthSet = true
	t.MaxLength = v
}

// Set reports whether any validations are set.
func (t Object) Set() bool {
	return t.MaxPropertiesSet || t.MinPropertiesSet
}

// ValidateProperties returns error if object length (properties number) v is invalid.
func (t Object) ValidateProperties(v int) error {
	if t.MaxPropertiesSet && v > t.MaxProperties {
		return errors.Errorf("object properties number %d greater than maximum %d", v, t.MaxProperties)
	}
	if t.MinPropertiesSet && v < t.MinProperties {
		return errors.Errorf("object properties number %d less than minimum %d", v, t.MinProperties)
	}
	if t.MinLengthSet && v < t.MinLength {
		return errors.Errorf("object length %d less than minimum %d", v, t.MinLength)
	}
	if t.MaxLengthSet && v > t.MaxLength {
		return errors.Errorf("object length number %d greater than maximum %d", v, t.MaxLength)
	}

	return nil
}
