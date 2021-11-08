package json

import (
	"github.com/go-faster/jx"
	"github.com/google/uuid"
)

// DecodeUUID decodes UUID from json.
func DecodeUUID(i *jx.Decoder) (v uuid.UUID, err error) {
	s, err := i.Str()
	if err != nil {
		return v, err
	}
	return uuid.Parse(s)
}

// EncodeUUID encodes UUID to json.
func EncodeUUID(s *jx.Encoder, v uuid.UUID) {
	s.Str(v.String())
}
