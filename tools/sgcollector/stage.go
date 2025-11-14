package main

type Stage int

func (s Stage) OnlyCounter() bool {
	switch s {
	case NotImplemented, Good:
		return true
	default:
		return false
	}
}

const (
	InvalidYAML Stage = iota
	InvalidJSON
	Unmarshal
	Parse
	BuildIR
	BuildRouter
	Template
	Format
	NotImplemented
	Good
	Crash
	last
)

func (s Stage) String() string {
	r := [last]string{
		"invalidYAML",
		"invalidJSON",
		"unmarshal",
		"parse",
		"buildIR",
		"buildRouter",
		"template",
		"format",
		"notImplemented",
		"good",
		last - 1: "crash",
	}
	if int(s) >= len(r) {
		return ""
	}
	return r[s] // #nosec G602
}
