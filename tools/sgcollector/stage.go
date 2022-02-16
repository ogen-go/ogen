package main

type Stage int

const (
	InvalidYAML Stage = iota
	InvalidJSON
	Parse
	Build
	Template
	Format
	Crash
	last
)

func (s Stage) String() string {
	r := [7]string{
		"invalidYAML",
		"invalidJSON",
		"parse",
		"build",
		"template",
		"format",
		last - 1: "crash",
	}
	if int(s) >= len(r) {
		return ""
	}
	return r[s]
}
