package match

type Matcher interface {
	Match(string) bool
}

type Matchers []Matcher

func (p Matchers) Len() int {
	return len(p)
}

func (p Matchers) MatchAll(strs []string) bool {
	if p.Len() != len(strs) {
		return false
	}
	for i, part := range p {
		if !part.Match(strs[i]) {
			return false
		}
	}
	return true
}

func (p Matchers) Each(fn func(int, Matcher) bool) bool {
	for i, matcher := range p {
		if fn(i, matcher) {
			return true
		}
	}
	return false
}
