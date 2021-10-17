package json

import (
	"time"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"
)

func ReadDate(i *Iterator) (v time.Time, err error) {
	return time.Parse(dateLayout, i.ReadString())
}

func WriteDate(s *Stream, v time.Time) {
	s.WriteString(v.Format(dateLayout))
}

func ReadTime(i *Iterator) (v time.Time, err error) {
	return time.Parse(timeLayout, i.ReadString())
}

func WriteTime(s *Stream, v time.Time) {
	s.WriteString(v.Format(timeLayout))
}

func ReadDateTime(i *Iterator) (v time.Time, err error) {
	return time.Parse(time.RFC3339, i.ReadString())
}

func WriteDateTime(s *Stream, v time.Time) {
	s.WriteString(v.Format(time.RFC3339))
}

func ReadDuration(i *Iterator) (v time.Duration, err error) {
	return time.ParseDuration(i.ReadString())
}

func WriteDuration(s *Stream, v time.Duration) {
	s.WriteString(v.String())
}
