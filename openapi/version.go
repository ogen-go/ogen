package openapi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-faster/errors"
)

// Version represents OpenAPI version.
type Version struct {
	// Major is the major version number.
	Major int
	// Minor is the minor version number.
	Minor int
	// Patch is the patch version number.
	Patch int
}

// MarshalText implements encoding.TextMarshaler.
func (v *Version) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (v *Version) UnmarshalText(text []byte) error {
	var (
		r      = string(text)
		err    error
		target Version
	)
	switch n := strings.Split(r, "."); len(n) {
	case 3:
		target.Patch, err = strconv.Atoi(n[2])
		if err != nil {
			return errors.Wrap(err, "invalid patch version")
		}
		fallthrough
	case 2:
		target.Minor, err = strconv.Atoi(n[1])
		if err != nil {
			return errors.Wrap(err, "invalid minor version")
		}
		fallthrough
	case 1:
		major := n[0]
		if len(n) == 1 {
			major = r
		}
		target.Major, err = strconv.Atoi(major)
		if err != nil {
			return errors.Wrap(err, "invalid major version")
		}
	default:
		return errors.New("version must have format <major>.<minor>[.<patch>]")
	}
	*v = target
	return nil
}

// String returns the string representation of the version.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}
