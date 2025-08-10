package xmaps

// BuildSet builds a set from the given values.
func BuildSet[K comparable](s ...K) map[K]struct{} {
	r := make(map[K]struct{}, len(s))
	for _, v := range s {
		r[v] = struct{}{}
	}
	return r
}
