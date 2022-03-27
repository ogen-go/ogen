package jsonpointer

import "strings"

func splitFunc(s string, sep byte, cb func(s string) error) error {
	for {
		idx := strings.IndexByte(s, sep)
		if idx < 0 {
			break
		}
		if err := cb(s[:idx]); err != nil {
			return err
		}
		s = s[idx+1:]
	}
	return cb(s)
}
