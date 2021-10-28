package json

import (
	"time"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"
)

func ReadDate(i *Iter) (v time.Time, err error) {
	return time.Parse(dateLayout, i.Str())
}

func WriteDate(s *Stream, v time.Time) {
	s.WriteString(v.Format(dateLayout))
}

func ReadTime(i *Iter) (v time.Time, err error) {
	return time.Parse(timeLayout, i.Str())
}

func WriteTime(s *Stream, v time.Time) {
	s.WriteString(v.Format(timeLayout))
}

func ReadDateTime(i *Iter) (v time.Time, err error) {
	return time.Parse(time.RFC3339, i.Str())
}

func WriteDateTime(s *Stream, v time.Time) {
	s.WriteString(v.Format(time.RFC3339))
}

func ReadDuration(i *Iter) (v time.Duration, err error) {
	return time.ParseDuration(i.Str())
}

func WriteDuration(s *Stream, v time.Duration) {
	s.WriteString(v.String())
}
