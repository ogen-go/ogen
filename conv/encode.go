package conv

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func IntToString(v int) string { return strconv.Itoa(v) }

func Int32ToString(v int32) string { return strconv.Itoa(int(v)) }

func Int64ToString(v int64) string { return strconv.FormatInt(v, 10) }

func Float32ToString(v float32) string { return strconv.FormatFloat(float64(v), 'f', 10, 64) }

func Float64ToString(v float64) string { return strconv.FormatFloat(v, 'f', 10, 64) }

func StringToString(v string) string { return v }

func BytesToString(v []byte) string { return string(v) }

func TimeToString(v time.Time) string { return v.Format(time.RFC3339) }

func BoolToString(v bool) string { return strconv.FormatBool(v) }

func UUIDToString(v uuid.UUID) string { return v.String() }

func InterfaceToString(v interface{}) string { return fmt.Sprintf("%s", v) }

func Int32ArrayToString(vs []int32) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, Int32ToString(v))
	}
	return strs
}

func Int64ArrayToString(vs []int64) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, Int64ToString(v))
	}
	return strs
}

func Float32ArrayToString(vs []float32) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, Float32ToString(v))
	}
	return strs
}

func Float64ArrayToString(vs []float64) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, Float64ToString(v))
	}
	return strs
}

func StringArrayToString(vs []string) []string {
	return vs
}

func BytesArrayToString(vs [][]byte) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, BytesToString(v))
	}
	return strs
}

func TimeArrayToString(vs []time.Time) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, TimeToString(v))
	}
	return strs
}

func BoolArrayToString(vs []bool) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, BoolToString(v))
	}
	return strs
}

func UUIDArrayToString(vs []uuid.UUID) []string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, UUIDToString(v))
	}
	return strs
}
