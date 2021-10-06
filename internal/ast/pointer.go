package ast

type Pointer struct {
	To Type
}

func (p *Pointer) Type() string {
	return "*" + p.To.Type()
}
