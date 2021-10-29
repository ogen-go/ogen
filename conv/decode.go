package conv

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

func ToInt(s string) (int, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int(v), err
}

func ToInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

func ToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func ToFloat32(s string) (float32, error) {
	v, err := strconv.ParseFloat(s, 32)
	return float32(v), err
}

func ToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func ToString(s string) (string, error) {
	return s, nil
}

func ToBytes(s string) ([]byte, error) {
	return []byte(s), nil
}

func ToTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func ToBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

func ToUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func ToInt32Array(a []string) ([]int32, error) {
	var arr []int32
	for _, s := range a {
		v, err := ToInt32(s)
		if err != nil {
			return nil, err
		}

		arr = append(arr, v)
	}

	return arr, nil
}

func ToInt64Array(a []string) ([]int64, error) {
	var arr []int64
	for _, s := range a {
		v, err := ToInt64(s)
		if err != nil {
			return nil, err
		}

		arr = append(arr, v)
	}

	return arr, nil
}

func ToFloat32Array(a []string) ([]float32, error) {
	var arr []float32
	for _, s := range a {
		v, err := ToFloat32(s)
		if err != nil {
			return nil, err
		}

		arr = append(arr, v)
	}

	return arr, nil
}

func ToFloat64Array(a []string) ([]float64, error) {
	var arr []float64
	for _, s := range a {
		v, err := ToFloat64(s)
		if err != nil {
			return nil, err
		}

		arr = append(arr, v)
	}

	return arr, nil
}

func ToStringArray(a []string) ([]string, error) {
	return a, nil
}

func ToBytesArray(a []string) ([][]byte, error) {
	var arr [][]byte
	for _, s := range a {

		arr = append(arr, []byte(s))
	}

	return arr, nil
}

func ToTimeArray(a []string) ([]time.Time, error) {
	var arr []time.Time
	for _, s := range a {
		v, err := ToTime(s)
		if err != nil {
			return nil, err
		}

		arr = append(arr, v)
	}

	return arr, nil
}

func ToBoolArray(a []string) ([]bool, error) {
	var arr []bool
	for _, s := range a {
		v, err := ToBool(s)
		if err != nil {
			return nil, err
		}

		arr = append(arr, v)
	}

	return arr, nil
}

func ToUUIDArray(a []string) ([]uuid.UUID, error) {
	var arr []uuid.UUID
	for _, s := range a {
		v, err := ToUUID(s)
		if err != nil {
			return nil, err
		}

		arr = append(arr, v)
	}

	return arr, nil
}
