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

type pathGroupDef struct {
	Path    string
	Methods []pathMethodDef
}

type pathMethodDef struct {
	Method string
}

type serverDef struct {
	Methods []serverMethodDef
}

type serverMethodDef struct {
	Name string
}
