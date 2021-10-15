package ast

type Pointer struct {
	To Type
}

func (p *Pointer) Type() string {
	return "*" + p.To.Type()
}

func (p *Pointer) needValidation(visited map[*Schema]struct{}) bool {
	switch to := p.To.(type) {
	case *Schema:
		return to.needValidation(visited)
	default:
		panic("unreachable")
	}
}
