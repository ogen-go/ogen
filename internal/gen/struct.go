package gen

type componentStructDef struct {
	Name        string
	Fields      []field
	Description string
	Path        string
}

type field struct {
	Name    string
	Type    string
	TagName string
}

type serverDef struct {
	Methods []serverMethodDef
}

type serverMethodDef struct {
	Name         string
	OperationID  string
	Path         string
	HTTPMethod   string
	Parameters   map[ParameterType][]Parameter
	RequestType  string
	ResponseType string
}
