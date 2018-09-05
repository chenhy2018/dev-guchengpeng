package channels

func inStrings(s string, ss []string) bool {
	for _, s0 := range ss {
		if s == s0 {
			return true
		}
	}
	return false
}

func union(a, b []string) []string {
	ret := make([]string, len(a), len(a)+len(b))
	copy(ret, a)
	for _, sb := range b {
		if !inStrings(sb, a) {
			ret = append(ret, sb)
		}
	}
	return ret
}

func intersection(a, b []string) []string {
	ret := make([]string, 0, len(b))
	for _, sb := range b {
		if inStrings(sb, a) {
			ret = append(ret, sb)
		}
	}
	return ret
}

func difference(a, b []string) []string {
	if len(b) == 0 {
		return a
	}
	ret := make([]string, 0, len(a))
	for _, sa := range a {
		if !inStrings(sa, b) {
			ret = append(ret, sa)
		}
	}
	return ret
}
