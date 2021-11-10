package ir

import "sort"

func methods(ifaces map[*TypeInterface]struct{}) []string {
	if ifaces == nil {
		return nil
	}

	ms := make(map[string]struct{}, len(ifaces))
	for iface := range ifaces {
		for m := range iface.Methods {
			ms[m] = struct{}{}
		}
	}

	result := make([]string, 0, len(ms))
	for m := range ms {
		result = append(result, m)
	}

	sort.Strings(result)
	return result
}
