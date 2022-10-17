package uri

import "strings"

type Field struct {
	Name  string
	Value string
}

func encodeObject(kvSep, fieldSep rune, fields []Field) string {
	var sb strings.Builder
	for i, f := range fields {
		if i > 0 {
			sb.WriteRune(fieldSep)
		}
		sb.WriteString(f.Name)
		sb.WriteRune(kvSep)
		sb.WriteString(f.Value)
	}
	return sb.String()
}

func decodeObject(cur *cursor, kvSep, fieldSep rune, f func(field, value string) error) error {
	var (
		fname string
		field = true
	)

	for {
		until := fieldSep
		if field {
			until = kvSep
		}

		v, hasNext, err := cur.readValue(until)
		if err != nil {
			return err
		}

		if field {
			fname = v
			field = false
			continue
		}

		field = true
		if err := f(fname, v); err != nil {
			return err
		}

		if !hasNext {
			return nil
		}
	}
}
