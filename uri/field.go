package uri

type fieldEncoder struct {
	set   bool
	value string
}

func (e *fieldEncoder) Value(v string) error {
	if e.set {
		panic("value already set")
	}
	e.value = v
	e.set = true
	return nil
}

func (e *fieldEncoder) Array(_ func(Encoder) error) error {
	panic("nested arrays not allowed in path parameters")
}

func (e *fieldEncoder) Field(_ string, _ func(Encoder) error) error {
	panic("nested objects not allowed in path parameters")
}
