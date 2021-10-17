package json

import (
	"time"

	json "github.com/json-iterator/go"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"
)

func ReadDate(i *json.Iterator) (v time.Time, err error) {
	return time.Parse(dateLayout, i.ReadString())
}

func WriteDate(s *json.Stream, v time.Time) {
	s.WriteString(v.Format(dateLayout))
}

func ReadTime(i *json.Iterator) (v time.Time, err error) {
	return time.Parse(timeLayout, i.ReadString())
}

func WriteTime(s *json.Stream, v time.Time) {
	s.WriteString(v.Format(timeLayout))
}

func ReadDateTime(i *json.Iterator) (v time.Time, err error) {
	return time.Parse(time.RFC3339, i.ReadString())
}

func WriteDateTime(s *json.Stream, v time.Time) {
	s.WriteString(v.Format(time.RFC3339))
}

func ReadDuration(i *json.Iterator) (v time.Duration, err error) {
	return time.ParseDuration(i.ReadString())
}

func WriteDuration(s *json.Stream, v time.Duration) {
	s.WriteString(v.String())
}
