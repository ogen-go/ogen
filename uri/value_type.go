package uri

type vtyp string

func (v vtyp) String() string { return string(v) }

const (
	vtNotSet vtyp = "notSet"
	vtValue  vtyp = "value"
	vtArray  vtyp = "array"
	vtObject vtyp = "object"
)
