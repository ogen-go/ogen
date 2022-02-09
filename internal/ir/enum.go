package ir

type EnumVariant struct {
	Name  string
	Value interface{}
}

func (v *EnumVariant) ValueGo() string {
	return PrintGoValue(v.Value)
}
