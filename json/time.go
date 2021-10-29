package json

import (
	"time"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"
)

func ReadDate(i *Iter) (v time.Time, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.Parse(dateLayout, s)
}

func WriteDate(s *Stream, v time.Time) {
	s.WriteString(v.Format(dateLayout))
}

func ReadTime(i *Iter) (v time.Time, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.Parse(timeLayout, s)
}

func WriteTime(s *Stream, v time.Time) {
	s.WriteString(v.Format(timeLayout))
}

func ReadDateTime(i *Iter) (v time.Time, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.Parse(time.RFC3339, s)
}

func WriteDateTime(s *Stream, v time.Time) {
	s.WriteString(v.Format(time.RFC3339))
}

func ReadDuration(i *Iter) (v time.Duration, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return time.ParseDuration(s)
}

func WriteDuration(s *Stream, v time.Duration) {
	s.WriteString(v.String())
}
