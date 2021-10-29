package json

import (
	"github.com/google/uuid"
)

func ReadUUID(i *Iter) (v uuid.UUID, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return uuid.Parse(s)
}

func WriteUUID(s *Stream, v uuid.UUID) {
	s.WriteString(v.String())
}
