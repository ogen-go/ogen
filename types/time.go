package types

import (
	"encoding/json"
	"time"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"
)

type Date struct {
	time.Time
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

type Duration struct {
	time.Duration
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
