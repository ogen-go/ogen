package uri

type pathFieldEncoder struct {
	set   bool
	value string
}

func (e *pathFieldEncoder) Value(v string) error {
	if e.set {
		panic("value already set")
	}
	e.value = v
	e.set = true
	return nil
}

func (e *pathFieldEncoder) Array(_ func(Encoder) error) error {
	panic("nested arrays not allowed in path parameters")
}

func (e *pathFieldEncoder) Field(_ string, _ func(Encoder) error) error {
	panic("nested objects not allowed in path parameters")
}
