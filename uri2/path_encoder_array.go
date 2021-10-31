package uri

type pathArrayEncoder struct {
	set   bool
	items []string
}

func (e *pathArrayEncoder) Value(v string) error {
	e.set = true
	e.items = append(e.items, v)
	return nil
}

func (e *pathArrayEncoder) Array(_ func(Encoder) error) error {
	panic("nested arrays not allowed in path parameters")
}

func (e *pathArrayEncoder) Field(_ string, _ func(Encoder) error) error {
	panic("nested objects not allowed in path parameters")
}
