package conv

import (
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/go-faster/errors"
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

func ToIP(s string) (net.IP, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, errors.Errorf("invalid ip: %q", s)
	}
	return ip, nil
}

func ToURL(s string) (url.URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return url.URL{}, err
	}
	return *u, nil
}

func ToDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

func ToInt32Array(a []string) ([]int32, error) {
	arr := make([]int32, len(a))
	for i := range a {
		v, err := ToInt32(a[i])
		if err != nil {
			return nil, err
		}

		arr[i] = v
	}

	return arr, nil
}

func ToInt64Array(a []string) ([]int64, error) {
	arr := make([]int64, len(a))
	for i := range a {
		v, err := ToInt64(a[i])
		if err != nil {
			return nil, err
		}

		arr[i] = v
	}

	return arr, nil
}

func ToFloat32Array(a []string) ([]float32, error) {
	arr := make([]float32, len(a))
	for i := range a {
		v, err := ToFloat32(a[i])
		if err != nil {
			return nil, err
		}

		arr[i] = v
	}

	return arr, nil
}

func ToFloat64Array(a []string) ([]float64, error) {
	arr := make([]float64, len(a))
	for i := range a {
		v, err := ToFloat64(a[i])
		if err != nil {
			return nil, err
		}

		arr[i] = v
	}

	return arr, nil
}

func ToStringArray(a []string) ([]string, error) {
	return a, nil
}

func ToBytesArray(a []string) ([][]byte, error) {
	arr := make([][]byte, len(a))
	for i := range a {
		arr[i] = []byte(a[i])
	}

	return arr, nil
}

func ToTimeArray(a []string) ([]time.Time, error) {
	arr := make([]time.Time, len(a))
	for i := range a {
		v, err := ToTime(a[i])
		if err != nil {
			return nil, err
		}

		arr[i] = v
	}

	return arr, nil
}

func ToBoolArray(a []string) ([]bool, error) {
	arr := make([]bool, len(a))
	for i := range a {
		v, err := ToBool(a[i])
		if err != nil {
			return nil, err
		}

		arr[i] = v
	}

	return arr, nil
}

func ToUUIDArray(a []string) ([]uuid.UUID, error) {
	arr := make([]uuid.UUID, len(a))
	for i := range a {
		v, err := ToUUID(a[i])
		if err != nil {
			return nil, err
		}

		arr[i] = v
	}

	return arr, nil
}
