package template

import (
	"fmt"
	"strconv"
	"testing"
)

type ExMessage struct {
	ID      int
	Message []byte
}

type ExContext struct {
	Messages []ExMessage
}

func ExAppend(b []byte, data ExContext) ([]byte, error) {
	// Header.
	b = append(b, `!<!DOCTYPE html>
<html>
<head><title>Fortunes</title></head>
<body>
<table>
<tr><th>id</th><th>message</th></tr>`...)

	// Content.
	for _, v := range data.Messages {
		b = append(b, `<tr><td>`...)
		b = append(b, strconv.Itoa(v.ID)...)
		b = append(b, `</td><td>`...)
		b = append(b, v.Message...)
		b = append(b, `</td></tr>`...)
		b = append(b, '\n')
	}

	// Footer.
	b = append(b, `</table>
</body>
</html>`...)

	return b, nil
}

func BenchmarkBaseline(b *testing.B) {
	var data ExContext
	for i := 0; i < 10; i++ {
		data.Messages = append(data.Messages, ExMessage{
			ID:      i + 1000,
			Message: []byte(fmt.Sprintf("Some message number %d!", i+1000)),
		})
	}

	b.ReportAllocs()
	b.ResetTimer()

	var (
		body []byte
		err  error
	)
	for i := 0; i < b.N; i++ {
		if body, err = ExAppend(body[:0], data); err != nil {
			b.Fatal(err)
		}
	}
}
