package ir

type EnumVariant struct {
	Name  string
	Value any
}

func (v *EnumVariant) ValueGo() string {
	return PrintGoValue(v.Value)
}
