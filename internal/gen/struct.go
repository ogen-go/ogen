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
