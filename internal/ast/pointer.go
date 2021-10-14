package ast

type Pointer struct {
	To Type
}

func (p *Pointer) Type() string {
	return "*" + p.To.Type()
}

func (p *Pointer) NeedValidation() bool {
	switch to := p.To.(type) {
	case *Schema:
		return to.NeedValidation()
	default:
		panic("unreachable")
	}
}
