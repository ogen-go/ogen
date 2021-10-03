//go:build !go1.17
// +build !go1.17

package json

import (
	json "github.com/bytedance/sonic"
)

// Marshal value to json.
func Marshal(val interface{}) ([]byte, error) {
	return json.Marshal(val)
}

// Unmarshal value from json.
func Unmarshal(data []byte, val interface{}) error {
	return json.Unmarshal(data, val)
}
