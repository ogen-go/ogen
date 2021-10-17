package json

import (
	json "github.com/json-iterator/go"
)

type Iterator = json.Iterator

func NewIterator(cfg API) *Iterator {
	return json.NewIterator(cfg)
}
