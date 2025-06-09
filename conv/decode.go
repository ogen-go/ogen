package conv

import (
	"encoding"
	stdjson "encoding/json"
	"net"
	"net/netip"
	"net/url"
	"slices"
	"strconv"
	"time"
	"unsafe"

	"github.com/go-faster/jx"
	"github.com/google/uuid"

	"github.com/ogen-go/ogen/json"
)

func ToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func ToInt8(s string) (int8, error) {
	v, err := strconv.ParseInt(s, 10, 8)
	return int8(v), err
}

func ToInt16(s string) (int16, error) {
	v, err := strconv.ParseInt(s, 10, 16)
	return int16(v), err
}

func ToInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

func ToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func ToUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 0)
	return uint(v), err
}

func ToUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)
	return uint8(v), err
}

func ToUint16(s string) (uint16, error) {
	v, err := strconv.ParseUint(s, 10, 16)
	return uint16(v), err
}

func ToUint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	return uint32(v), err
}

func ToUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
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

// toBytes converts a string to a byte slice with zero allocation.
func toBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s)) //nolint:gosec // Unsafe conversion is intended for performance.
}

func ToBytes(s string) ([]byte, error) {
	return toBytes(s), nil
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

func ToMAC(s string) (net.HardwareAddr, error) {
	return net.ParseMAC(s)
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

func ToStringInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func ToStringInt8(s string) (int8, error) {
	v, err := strconv.ParseInt(s, 10, 8)
	return int8(v), err
}

func ToStringInt16(s string) (int16, error) {
	v, err := strconv.ParseInt(s, 10, 16)
	return int16(v), err
}

func ToStringInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

func ToStringInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func ToStringUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 0)
	return uint(v), err
}

func ToStringUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)
	return uint8(v), err
}

func ToStringUint16(s string) (uint16, error) {
	v, err := strconv.ParseUint(s, 10, 16)
	return uint16(v), err
}

func ToStringUint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	return uint32(v), err
}

func ToStringUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

func ToStringFloat32(s string) (float32, error) {
	v, err := strconv.ParseFloat(s, 32)
	return float32(v), err
}

func ToStringFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

type (
	ogenUnmarshaler[T any] interface {
		json.Unmarshaler
		*T
	}
	textUnmarshaler[T any] interface {
		encoding.TextUnmarshaler
		*T
	}
	jsonUnmarshaler[T any] interface {
		stdjson.Unmarshaler
		*T
	}
)

func ToNative[T any, P ogenUnmarshaler[T]](s string) (T, error) {
	var v T
	err := P(&v).Decode(jx.DecodeBytes(toBytes(s)))
	return v, err
}

func ToStringNative[T any, P ogenUnmarshaler[T]](s string) (T, error) {
	return ToNative[T, P](strconv.Quote(s))
}

func ToText[T any, P textUnmarshaler[T]](s string) (T, error) {
	var v T
	err := P(&v).UnmarshalText(toBytes(s))
	return v, err
}

func ToStringText[T any, P textUnmarshaler[T]](s string) (T, error) {
	return ToText[T, P](s)
}

func ToJSON[T any, P jsonUnmarshaler[T]](s string) (T, error) {
	var v T
	err := P(&v).UnmarshalJSON(toBytes(s))
	return v, err
}

func ToStringJSON[T any, P jsonUnmarshaler[T]](s string) (T, error) {
	return ToJSON[T, P](strconv.Quote(s))
}

func ToExternal[T any](s string) (T, error) {
	var v T
	err := stdjson.Unmarshal(toBytes(s), &v)
	return v, err
}

func ToStringExternal[T any](s string) (T, error) {
	return ToExternal[T](strconv.Quote(s))
}

func decodeArray[T any](a []string, decode func(string) (T, error)) ([]T, error) {
	arr := make([]T, len(a))
	for i := range a {
		v, err := decode(a[i])
		if err != nil {
			return nil, err
		}
		arr[i] = v
	}
	return arr, nil
}

func ToInt32Array(a []string) ([]int32, error) {
	return decodeArray(a, ToInt32)
}

func ToInt64Array(a []string) ([]int64, error) {
	return decodeArray(a, ToInt64)
}

func ToFloat32Array(a []string) ([]float32, error) {
	return decodeArray(a, ToFloat32)
}

func ToFloat64Array(a []string) ([]float64, error) {
	return decodeArray(a, ToFloat64)
}

func ToStringArray(a []string) ([]string, error) {
	return slices.Clone(a), nil
}

func ToBytesArray(a []string) ([][]byte, error) {
	return decodeArray(a, ToBytes)
}

func ToTimeArray(a []string) ([]time.Time, error) {
	return decodeArray(a, ToTime)
}

func ToBoolArray(a []string) ([]bool, error) {
	return decodeArray(a, ToBool)
}

func ToUUIDArray(a []string) ([]uuid.UUID, error) {
	return decodeArray(a, ToUUID)
}

func ToMACArray(a []string) ([]net.HardwareAddr, error) {
	return decodeArray(a, ToMAC)
}
