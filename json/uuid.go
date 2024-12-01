package json

import (
	"github.com/go-faster/jx"
	"github.com/google/uuid"
)

// DecodeUUID decodes UUID from json.
func DecodeUUID(i *jx.Decoder) (v uuid.UUID, err error) {
	s, err := i.StrBytes()
	if err != nil {
		return v, err
	}
	return uuid.ParseBytes(s)
}

// EncodeUUID encodes UUID to json.
func EncodeUUID(s *jx.Encoder, v uuid.UUID) {
	const (
		// Hexed length (16 * 2) + 4 hyphens
		length       = len(v)*2 + 4
		quotedLength = length + 2
	)
	dst := [quotedLength]byte{
		0:                '"',
		quotedLength - 1: '"',
	}
	hexEncode((*[length]byte)(dst[1:length+1]), v)
	s.Raw(dst[:])
}

func hexEncode(dst *[36]byte, v uuid.UUID) {
	const hextable = "0123456789abcdef"

	{
		dst[0], dst[1] = hextable[v[0]>>4], hextable[v[0]&0x0f]
		dst[2], dst[3] = hextable[v[1]>>4], hextable[v[1]&0x0f]
		dst[4], dst[5] = hextable[v[2]>>4], hextable[v[2]&0x0f]
		dst[6], dst[7] = hextable[v[3]>>4], hextable[v[3]&0x0f]
	}
	dst[8] = '-'
	{
		dst[9+0], dst[9+1] = hextable[v[4+0]>>4], hextable[v[4+0]&0x0f]
		dst[9+2], dst[9+3] = hextable[v[4+1]>>4], hextable[v[4+1]&0x0f]
	}
	dst[13] = '-'
	{
		dst[14+0], dst[14+1] = hextable[v[6+0]>>4], hextable[v[6+0]&0x0f]
		dst[14+2], dst[14+3] = hextable[v[6+1]>>4], hextable[v[6+1]&0x0f]
	}
	dst[18] = '-'
	{
		dst[19+0], dst[19+1] = hextable[v[8+0]>>4], hextable[v[8+0]&0x0f]
		dst[19+2], dst[19+3] = hextable[v[8+1]>>4], hextable[v[8+1]&0x0f]
	}
	dst[23] = '-'
	{
		dst[24+0], dst[24+1] = hextable[v[10+0]>>4], hextable[v[10+0]&0x0f]
		dst[24+2], dst[24+3] = hextable[v[10+1]>>4], hextable[v[10+1]&0x0f]
		dst[24+4], dst[24+5] = hextable[v[10+2]>>4], hextable[v[10+2]&0x0f]
		dst[24+6], dst[24+7] = hextable[v[10+3]>>4], hextable[v[10+3]&0x0f]
		dst[24+8], dst[24+9] = hextable[v[10+4]>>4], hextable[v[10+4]&0x0f]
		dst[24+10], dst[24+11] = hextable[v[10+5]>>4], hextable[v[10+5]&0x0f]
	}
}
