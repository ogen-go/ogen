package gen

// config is input for code generation templates.
type config struct {
	Package string
	Schemas []schemaStructDef
	Server  serverDef
}
