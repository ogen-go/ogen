package json

import (
	"github.com/google/uuid"
)

func ReadUUID(i *Iterator) (v uuid.UUID, err error) {
	return uuid.Parse(i.ReadString())
}

func WriteUUID(s *Stream, v uuid.UUID) {
	s.WriteString(v.String())
}
