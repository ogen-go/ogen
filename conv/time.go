package conv

import (
	"time"
)

const (
	dateLayout     = "2006-01-02"
	timeLayout     = "15:04:05"
	httpDateLayout = "Mon, 02 Jan 2006 15:04:05 GMT"
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

func HTTPDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
}
