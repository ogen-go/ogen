package json

import (
	json "github.com/json-iterator/go"
)

type Iterator = json.Iterator

func NewIterator() *Iterator {
	return json.NewIterator(ConfigDefault)
}
