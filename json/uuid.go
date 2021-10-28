package json

import (
	"github.com/google/uuid"
)

func ReadUUID(i *Iter) (v uuid.UUID, err error) {
	return uuid.Parse(i.Str())
}

func WriteUUID(s *Stream, v uuid.UUID) {
	s.WriteString(v.String())
}
