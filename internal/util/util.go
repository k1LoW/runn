package util

// Unique returns a slice with duplicates removed while maintaining order.
func Unique(in []string) []string {
	var u []string
	m := map[string]struct{}{}
	for _, s := range in {
		if _, ok := m[s]; ok {
			continue
		}
		u = append(u, s)
		m[s] = struct{}{}
	}
	return u
}
