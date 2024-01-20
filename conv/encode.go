package conv

import (
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func IntToString(v int) string     { return strconv.Itoa(v) }
func Int8ToString(v int8) string   { return strconv.FormatInt(int64(v), 10) }
func Int16ToString(v int16) string { return strconv.FormatInt(int64(v), 10) }
func Int32ToString(v int32) string { return strconv.FormatInt(int64(v), 10) }
func Int64ToString(v int64) string { return strconv.FormatInt(v, 10) }

func UintToString(v uint) string     { return strconv.FormatUint(uint64(v), 10) }
func Uint8ToString(v uint8) string   { return strconv.FormatUint(uint64(v), 10) }
func Uint16ToString(v uint16) string { return strconv.FormatUint(uint64(v), 10) }
func Uint32ToString(v uint32) string { return strconv.FormatUint(uint64(v), 10) }
func Uint64ToString(v uint64) string { return strconv.FormatUint(v, 10) }

func Float32ToString(v float32) string { return strconv.FormatFloat(float64(v), 'f', 10, 64) }
func Float64ToString(v float64) string { return strconv.FormatFloat(v, 'f', 10, 64) }

func BoolToString(v bool) string { return strconv.FormatBool(v) }

func StringToString(v string) string { return v }
func BytesToString(v []byte) string  { return string(v) }

func TimeToString(v time.Time) string     { return v.Format(timeLayout) }
func DateToString(v time.Time) string     { return v.Format(dateLayout) }
func DateTimeToString(v time.Time) string { return v.Format(time.RFC3339) }

func UnixSecondsToString(v time.Time) string { return StringInt64ToString(v.Unix()) }
func UnixNanoToString(v time.Time) string    { return StringInt64ToString(v.UnixNano()) }
func UnixMicroToString(v time.Time) string   { return StringInt64ToString(v.UnixMicro()) }
func UnixMilliToString(v time.Time) string   { return StringInt64ToString(v.UnixMilli()) }

func DurationToString(v time.Duration) string { return v.String() }

func UUIDToString(v uuid.UUID) string { return v.String() }

func MACToString(v net.HardwareAddr) string { return v.String() }

func AddrToString(v netip.Addr) string { return v.String() }

func URLToString(v url.URL) string { return v.String() }

func StringIntToString(v int) string     { return strconv.FormatInt(int64(v), 10) }
func StringInt8ToString(v int8) string   { return strconv.FormatInt(int64(v), 10) }
func StringInt16ToString(v int16) string { return strconv.FormatInt(int64(v), 10) }
func StringInt32ToString(v int32) string { return strconv.FormatInt(int64(v), 10) }
func StringInt64ToString(v int64) string { return strconv.FormatInt(v, 10) }

func StringUintToString(v uint) string     { return strconv.FormatUint(uint64(v), 10) }
func StringUint8ToString(v uint8) string   { return strconv.FormatUint(uint64(v), 10) }
func StringUint16ToString(v uint16) string { return strconv.FormatUint(uint64(v), 10) }
func StringUint32ToString(v uint32) string { return strconv.FormatUint(uint64(v), 10) }
func StringUint64ToString(v uint64) string { return strconv.FormatUint(v, 10) }

func StringFloat32ToString(v float32) string { return strconv.FormatFloat(float64(v), 'g', 10, 32) }
func StringFloat64ToString(v float64) string { return strconv.FormatFloat(v, 'g', 10, 64) }

func encodeArray[T any](vs []T, encode func(T) string) []string {
	strs := make([]string, len(vs))
	for i, v := range vs {
		strs[i] = encode(v)
	}
	return strs
}

func Int32ArrayToString(vs []int32) []string {
	return encodeArray(vs, Int32ToString)
}

func Int64ArrayToString(vs []int64) []string {
	return encodeArray(vs, Int64ToString)
}

func Float32ArrayToString(vs []float32) []string {
	return encodeArray(vs, Float32ToString)
}

func Float64ArrayToString(vs []float64) []string {
	return encodeArray(vs, Float64ToString)
}

func StringArrayToString(vs []string) []string {
	return vs
}

func BytesArrayToString(vs [][]byte) []string {
	return encodeArray(vs, BytesToString)
}

func TimeArrayToString(vs []time.Time) []string {
	return encodeArray(vs, TimeToString)
}

func BoolArrayToString(vs []bool) []string {
	return encodeArray(vs, BoolToString)
}

func UUIDArrayToString(vs []uuid.UUID) []string {
	return encodeArray(vs, UUIDToString)
}

func MACArrayToString(vs []net.HardwareAddr) []string {
	return encodeArray(vs, MACToString)
}
