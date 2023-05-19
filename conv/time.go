package conv

import (
	"time"
)

func Date(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func Time(t time.Time) time.Time {
	return time.Date(0, 0, 0, t.Hour(), t.Minute(), t.Second(), 0, t.Location())
}

func DateTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
}
