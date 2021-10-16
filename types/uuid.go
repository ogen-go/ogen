package types

import (
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

func ReadUUID(i *jsoniter.Iterator) (v uuid.UUID, err error) {
	if err := v.UnmarshalText([]byte(i.ReadString())); err != nil {
		return v, err
	}
	return v, nil
}

func WriteUUID(s *jsoniter.Stream, v uuid.UUID) {
	b, _ := v.MarshalText()
	s.WriteString(string(b))
}
