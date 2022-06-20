package conv

import (
	"net/netip"
	"net/url"
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
	return time.Parse(timeLayout, s)
}

func ToDate(s string) (time.Time, error) {
	return time.Parse(dateLayout, s)
}

func ToDateTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func ToUnixSeconds(s string) (time.Time, error) {
	val, err := ToInt64(s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(val, 0), nil
}

func ToUnixNano(s string) (time.Time, error) {
	val, err := ToInt64(s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, val), nil
}

func ToUnixMicro(s string) (time.Time, error) {
	val, err := ToInt64(s)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMicro(val), nil
}

func ToUnixMilli(s string) (time.Time, error) {
	val, err := ToInt64(s)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(val), nil
}

func ToBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

func ToUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func ToAddr(s string) (netip.Addr, error) {
	return netip.ParseAddr(s)
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

func ToStringInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

func ToStringInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
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
