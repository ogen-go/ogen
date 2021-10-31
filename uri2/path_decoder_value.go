package uri

type pathDecoderValue struct {
	read bool
	v    string
}

func (d *pathDecoderValue) Value() (string, error) {
	if d.read {
		panic("multiple Value calls")
	}

	d.read = true
	return d.v, nil
}

func (d *pathDecoderValue) Array(f func(Decoder) error) error {
	panic("nested arrays not allowed")
}

func (d *pathDecoderValue) Fields(f func(string, Decoder) error) error {
	panic("nested objects not allowed")
}
