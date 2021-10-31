package uri

type constval struct {
	v string
}

func (v constval) Value() (string, error) {
	return v.v, nil
}

func (d constval) Array(f func(Decoder) error) error {
	panic("its a value, not an array")
}

func (d constval) Fields(f func(string, Decoder) error) error {
	panic("its a value, not an object")
}
