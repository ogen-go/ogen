package json

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ogen-go/jx"
	"github.com/stretchr/testify/require"
)

// World represents modified WorldObject of TechEmpower OAS.
//
// Using as reference simple json object.
type World struct {
	ID           int64  `json:"id"`
	RandomNumber int64  `json:"randomNumber"`
	Message      string `json:"message"`
}

func (w World) String() string {
	return fmt.Sprintf("id=%d n=%d msg=%s", w.ID, w.RandomNumber, w.Message)
}

type WorldData struct {
	Value World
	Raw   RawMessage
	Len   int
}

func (d WorldData) Bytes() []byte {
	return append([]byte{}, d.Raw...)
}

func (d WorldData) Setup(b *testing.B) {
	b.Helper()

	b.ResetTimer()
	b.SetBytes(int64(d.Len))
	b.ReportAllocs()
}

func testWorld(t testing.TB) WorldData {
	t.Helper()

	v := World{
		ID:           10,
		RandomNumber: 12351,
		Message:      "Hello, world!",
	}

	data, err := Marshal(v)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	t.Logf("Payload: %s", v)
	t.Logf("Payload raw: %s", data)
	t.Logf("Payload size: %d", len(data))

	return WorldData{
		Value: v,
		Raw:   data,
		Len:   len(data),
	}
}

func BenchmarkMarshal(b *testing.B) {
	b.Run("World", func(b *testing.B) {
		d := testWorld(b)

		b.Run("std", func(b *testing.B) {
			d.Setup(b)

			for i := 0; i < b.N; i++ {
				data, err := Marshal(d.Value)
				require.NoError(b, err)
				require.NotEmpty(b, data)
			}
		})
		b.Run("jsoniter", func(b *testing.B) {
			b.Run("Encoder", func(b *testing.B) {
				d.Setup(b)

				e := GetEncoder()
				defer PutEncoder(e)

				for i := 0; i < b.N; i++ {
					e.Reset()
					e.ObjStart()

					// "id": 10,
					e.ObjField("id")
					e.Int64(d.Value.ID)
					e.More()

					// "randomNumber": 12351,
					e.ObjField("randomNumber")
					e.Int64(d.Value.RandomNumber)
					e.More()

					// "message": "Hello, world!"
					e.ObjField("message")
					e.Str(d.Value.Message)

					e.ObjEnd()
				}
			})
		})

	})
}

func BenchmarkUnmarshal(b *testing.B) {
	b.Run("World", func(b *testing.B) {
		d := testWorld(b)

		b.Run("std", func(b *testing.B) {
			d.Setup(b)

			data := d.Bytes()

			for i := 0; i < b.N; i++ {
				var v World

				if err := Unmarshal(data, &v); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run("jsoniter", func(b *testing.B) {
			d.Setup(b)

			r := jx.GetDecoder()
			data := d.Bytes()

			for i := 0; i < b.N; i++ {
				r.ResetBytes(data)

				var v World
				if err := r.ObjBytes(func(r *Decoder, k []byte) error {
					switch string(k) {
					case "id":
						n, err := r.Int64()
						if err != nil {
							return err
						}
						v.ID = n
					case "randomNumber":
						n, err := r.Int64()
						if err != nil {
							return err
						}
						v.RandomNumber = n
					case "message":
						s, err := r.Str()
						if err != nil {
							return err
						}
						v.Message = s
					default:
						return errors.New("unexpected key")
					}
					return nil
				}); err != nil {
					b.Fatal(err)
				}

				if v.Message == "" || v.ID == 0 || v.RandomNumber == 0 {
					b.Errorf("bad read: %s", v)
				}
			}
		})
	})
}
