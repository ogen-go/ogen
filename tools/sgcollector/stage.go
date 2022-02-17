package main

type Stage int

const (
	InvalidYAML Stage = iota
	InvalidJSON
	Unmarshal
	Parse
	BuildIR
	BuildRouter
	Template
	Format
	Crash
	last
)

func (s Stage) String() string {
	r := [9]string{
		"invalidYAML",
		"invalidJSON",
		"unmarshal",
		"parse",
		"buildIR",
		"buildRouter",
		"template",
		"format",
		last - 1: "crash",
	}
	if int(s) >= len(r) {
		return ""
	}
	return r[s]
}
