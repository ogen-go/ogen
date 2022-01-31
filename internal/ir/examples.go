package ir

import "encoding/json"

func (t *Type) Examples() (r []json.RawMessage) {
	if t.Schema == nil {
		return nil
	}

	dedup := make(map[string]struct{}, len(t.Schema.Examples))
	for _, example := range t.Schema.Examples {
		if !json.Valid(example) {
			continue
		}
		if _, ok := dedup[string(example)]; ok {
			continue
		}
		dedup[string(example)] = struct{}{}
		r = append(r, example)
	}

	return
}
