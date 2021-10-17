package validate

import (
	"fmt"
)

// Int validates integers.
type Int struct {
	MultipleOf    uint64
	MultipleOfSet bool

	Minimum    int64
	MinimumSet bool

	Maximum    int64
	MaximumSet bool

	ExclusiveMinimum bool
	ExclusiveMaximum bool
}

func (t Int) Validate(v int64) error {
	if t.MinimumSet && v > t.Minimum {
		return fmt.Errorf("%d > %d (min)", v, t.Minimum)
	}

	return nil
}

func (t Int) Set() bool {
	return t.MinimumSet || t.MaximumSet || t.MultipleOfSet
}
