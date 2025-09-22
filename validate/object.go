package validate

import "github.com/go-faster/errors"

// Object validates map length.
type Object struct {
	MinProperties    int
	MinPropertiesSet bool
	MaxProperties    int
	MaxPropertiesSet bool
	// TODO: add validate to gen
	MinLength    int
	MinLengthSet bool
	MaxLength    int
	MaxLengthSet bool
}

// SetMinProperties sets MinProperties validation.
func (o *Object) SetMinProperties(v int) {
	o.MinPropertiesSet = true
	o.MinProperties = v
}

// SetMaxProperties sets MaxProperties validation.
func (o *Object) SetMaxProperties(v int) {
	o.MaxPropertiesSet = true
	o.MaxProperties = v
}

// SetMinLength sets MinLength validation.
func (o *Object) SetMinLength(v int) {
	o.MinLengthSet = true
	o.MinLength = v
}

// SetMaxLength sets MaxLength validation.
func (o *Object) SetMaxLength(v int) {
	o.MaxLengthSet = true
	o.MaxLength = v
}

// Set reports whether any validations are seo.
func (o Object) Set() bool {
	return o.MaxPropertiesSet || o.MinPropertiesSet
}

// ValidateProperties returns error if object length (properties number) v is invalid.
func (o Object) ValidateProperties(v int) error {
	if o.MaxPropertiesSet && v > o.MaxProperties {
		return errors.Errorf("object properties number %d greater than maximum %d", v, o.MaxProperties)
	}
	if o.MinPropertiesSet && v < o.MinProperties {
		return errors.Errorf("object properties number %d less than minimum %d", v, o.MinProperties)
	}

	return nil
}
