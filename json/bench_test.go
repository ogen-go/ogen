package json

import (
	"fmt"
	"testing"

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

type bufferWriter struct {
	Data []byte
}

func (b *bufferWriter) Reset() {
	b.Data = b.Data[:0]
}

func (b *bufferWriter) Write(p []byte) (n int, err error) {
	b.Data = append(b.Data, p...)
	return len(p), nil
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
			b.Run("Stream", func(b *testing.B) {
				d.Setup(b)

				var w bufferWriter
				s := NewStream(&w)

				for i := 0; i < b.N; i++ {
					s.WriteObjectStart()

					// "id": 10,
					s.WriteObjectField("id")
					s.WriteInt64(d.Value.ID)
					s.WriteMore()

					// "randomNumber": 12351,
					s.WriteObjectField("randomNumber")
					s.WriteInt64(d.Value.RandomNumber)
					s.WriteMore()

					// "message": "Hello, world!"
					s.WriteObjectField("message")
					s.WriteString(d.Value.Message)

					s.WriteObjectEnd()

					if err := s.Flush(); err != nil {
						b.Fatal(err)
					}

					w.Reset()
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

			iter := newIter()
			data := d.Bytes()

			for i := 0; i < b.N; i++ {
				iter.ResetBytes(data)

				var v World
				iter.Object(func(iter *Iter, k string) bool {
					switch k {
					case "id":
						v.ID = iter.Int64()
					case "randomNumber":
						v.RandomNumber = iter.Int64()
					case "message":
						v.Message = iter.Str()
					default:
						b.Errorf("unexpected key %s", k)
						return false
					}
					return true
				})

				if v.Message == "" || v.ID == 0 || v.RandomNumber == 0 {
					b.Errorf("bad read: %s", v)
				}
			}
		})
	})
}
