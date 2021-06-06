package gen

// config is input for code generation templates.
type config struct {
	Package    string
	Components []componentStructDef
	Groups     []pathGroupDef
	Server     serverDef
}
