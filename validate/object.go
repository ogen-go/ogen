package validate

import "github.com/go-faster/errors"

// Object validates map length.
type Object struct {
	MinProperties    int
	MinPropertiesSet bool
	MaxProperties    int
	MaxPropertiesSet bool
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

	return nil
}
