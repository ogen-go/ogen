package json

import (
	"github.com/google/uuid"
	json "github.com/json-iterator/go"
)

func ReadUUID(i *json.Iterator) (v uuid.UUID, err error) {
	return uuid.Parse(i.ReadString())
}

func WriteUUID(s *json.Stream, v uuid.UUID) {
	s.WriteString(v.String())
}
