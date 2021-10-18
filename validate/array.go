package validate

import "fmt"

type Array struct {
	MinLength    int
	MinLengthSet bool
	MaxLength    int
	MaxLengthSet bool
}

func (t *Array) SetMaxLength(v int) {
	t.MaxLengthSet = true
	t.MaxLength = v
}

func (t *Array) SetMinLength(v int) {
	t.MinLengthSet = true
	t.MinLength = v
}

func (t Array) ValidateLength(v int) error {
	if t.MaxLengthSet && v > t.MaxLength {
		return fmt.Errorf("max: %d", t.MaxLength)
	}
	if t.MinLengthSet && v < t.MinLength {
		return fmt.Errorf("min: %d", t.MinLength)
	}

	return nil
}

func (t Array) Set() bool {
	return t.MaxLengthSet || t.MinLengthSet
}
