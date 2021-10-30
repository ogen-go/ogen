package json

import (
	"github.com/google/uuid"
)

func ReadUUID(i *Decoder) (v uuid.UUID, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return uuid.Parse(s)
}

func WriteUUID(s *Encoder, v uuid.UUID) {
	s.Str(v.String())
}
