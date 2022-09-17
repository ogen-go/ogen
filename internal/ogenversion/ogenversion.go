// Package ogenversion provides the version of the ogen tool.
package ogenversion

import (
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

var getOnce struct {
	info Info
	ok   bool
	once sync.Once
}

func getOgenVersion(m *debug.Module) (string, bool) {
	if m == nil || m.Path != "github.com/ogen-go/ogen" {
		return "", false
	}
	return m.Version, true
}

func getInfo() (Info, bool) {
	getOnce.once.Do(func() {
		bi, ok := debug.ReadBuildInfo()
		getOnce.ok = ok
		if !ok {
			return
		}

		var ogenIsDep bool
		if v, ok := getOgenVersion(&bi.Main); ok {
			getOnce.info.Version = v
		} else {
			ogenIsDep = true
			for _, m := range bi.Deps {
				if v, ok := getOgenVersion(m); ok {
					getOnce.info.Version = v
					break
				}
			}
		}
		getOnce.info.GoVersion = bi.GoVersion
		if !ogenIsDep {
			// ogen is the main module, so we can use buildvcs data.
			for _, s := range bi.Settings {
				switch s.Key {
				case "vcs.revision":
					getOnce.info.Commit = s.Value
				case "vcs.time":
					if t, err := time.Parse(time.RFC3339Nano, s.Value); err == nil {
						getOnce.info.Time = t
					}
				}
			}
		}
	})
	return getOnce.info, getOnce.ok
}

// Info is the ogen build information.
type Info struct {
	// Version is the version of the ogen tool.
	Version string
	// GoVersion is the version of the Go that produced this binary.
	GoVersion string

	// Commit is the current commit hash.
	Commit string
	// Time is the time of the build.
	Time time.Time
}

// GetInfo returns the ogen build information.
func GetInfo() (Info, bool) {
	return getInfo()
}

// String returns string representation of the build information.
func (i Info) String() string {
	var s strings.Builder
	s.WriteString("ogen version ")
	if v := i.Version; v != "" {
		s.WriteString(v)
	} else {
		s.WriteString("unknown")
	}
	if commit := i.Commit; commit != "" {
		s.WriteByte('-')
		s.WriteString(commit)
	}

	if t, v := i.Time, i.GoVersion; v != "" || !t.IsZero() {
		s.WriteString(" (built")
		if v != "" {
			s.WriteString(" with ")
			s.WriteString(v)
		}
		if !t.IsZero() {
			s.WriteString(" at ")
			s.WriteString(t.Format(time.RFC1123))
		}
		s.WriteByte(')')
	}
	const osArch = " " + runtime.GOOS + "/" + runtime.GOARCH
	s.WriteString(osArch)
	return s.String()
}
