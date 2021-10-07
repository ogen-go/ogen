package ast

type Interface struct {
	Name            string
	Methods         map[string]struct{}
	Implementations map[*Schema]struct{}
}

func (i *Interface) AddMethod(method string) {
	if i.Implementations == nil {
		i.Implementations = map[*Schema]struct{}{}
	}
	if i.Methods == nil {
		i.Methods = map[string]struct{}{}
	}
	i.Methods[method] = struct{}{}
}

func (i *Interface) Type() string { return i.Name }
