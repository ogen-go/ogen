package types

import (
	"encoding/json"
	"time"

	jsoniter "github.com/json-iterator/go"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"
)

type Date struct {
	time.Time
}

func ReadDate(i *jsoniter.Iterator) (v time.Time, err error) {
	return time.Parse(dateLayout, i.ReadString())
}

func WriteDate(s *jsoniter.Stream, v time.Time) {
	s.WriteString(v.Format(dateLayout))
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(dateLayout))
}

func (d *Date) UnmarshalJSON(b []byte) error {
	var v string

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	t, err := time.Parse(dateLayout, v)
	if err != nil {
		return err
	}

	d.Time = t

	return nil
}

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Format(timeLayout))
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var v string

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	dt, err := time.Parse(timeLayout, v)
	if err != nil {
		return err
	}

	t.Time = dt

	return nil
}

func ReadTime(i *jsoniter.Iterator) (v time.Time, err error) {
	return time.Parse(timeLayout, i.ReadString())
}

func WriteTime(s *jsoniter.Stream, v time.Time) {
	s.WriteString(v.Format(timeLayout))
}

func ReadDateTime(i *jsoniter.Iterator) (v time.Time, err error) {
	return time.Parse(time.RFC3339, i.ReadString())
}

func WriteDateTime(s *jsoniter.Stream, v time.Time) {
	s.WriteString(v.Format(time.RFC3339))
}

type Duration struct {
	time.Duration
}

func ReadDuration(i *jsoniter.Iterator) (v time.Duration, err error) {
	return time.ParseDuration(i.ReadString())
}

func WriteDuration(s *jsoniter.Stream, v time.Duration) {
	s.WriteString(v.String())
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v string

	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	dt, err := time.ParseDuration(v)
	if err != nil {
		return err
	}

	d.Duration = dt

	return nil
}
