package uri

type arrayEncoder struct {
	set   bool
	items []string
}

func (e *arrayEncoder) Value(v string) error {
	e.set = true
	e.items = append(e.items, v)
	return nil
}

func (e *arrayEncoder) Array(_ func(Encoder) error) error {
	panic("nested arrays not allowed in path parameters")
}

func (e *arrayEncoder) Field(_ string, _ func(Encoder) error) error {
	panic("nested objects not allowed in path parameters")
}
