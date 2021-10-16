package validate

type String struct {
	MinLength    int
	MinLengthSet bool
	MaxLength    int
	MaxLengthSet bool
}

func (t *String) SetMaxLength(v int) {
	t.MaxLengthSet = true
	t.MaxLength = v
}

func (t *String) SetMinLength(v int) {
	t.MinLengthSet = true
	t.MinLength = v
}

func (t String) Set() bool {
	return t.MaxLengthSet || t.MinLengthSet
}

func (t String) Validate(v string) error {
	return Array(t).ValidateLength(len([]rune(v)))
}
