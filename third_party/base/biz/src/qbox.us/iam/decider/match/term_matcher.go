package match

import (
	"strings"
)

type termMatchType uint8

const (
	termMatchTypeFull    termMatchType = 0x01
	termMatchTypePrefix  termMatchType = 0x02
	termMatchTypeSuffix  termMatchType = 0x04
	termMatchTypeContain termMatchType = termMatchTypePrefix | termMatchTypeSuffix
)

type TermMatcher struct {
	matchType   termMatchType
	matchStr    string
	originalStr string
}

func NewTermMatcher(str string) (matcher *TermMatcher) {
	matcher = &TermMatcher{originalStr: str, matchStr: str}

	if str == "*" || str == "" {
		matcher.matchType = termMatchTypeContain
		matcher.matchStr = ""
		return
	}
Loop:
	for {
		idx := strings.IndexByte(matcher.matchStr, '*')
		if idx == -1 {
			break
		}
		switch idx {
		case len(matcher.matchStr) - 1:
			matcher.matchType |= termMatchTypePrefix
			matcher.matchStr = matcher.matchStr[:idx]
		case 0:
			matcher.matchType |= termMatchTypeSuffix
			matcher.matchStr = matcher.matchStr[idx+1:]
		default:
			break Loop
		}
	}
	if matcher.matchType == 0 {
		matcher.matchType = termMatchTypeFull
	}
	return
}

func (m *TermMatcher) IsFullMatch() bool {
	return m.matchType == termMatchTypeFull
}

func (m *TermMatcher) IsContainMatch() bool {
	return m.matchType == termMatchTypeContain
}

func (m *TermMatcher) IsPrefixMatch() bool {
	return m.matchType == termMatchTypePrefix
}

func (m *TermMatcher) IsSuffixMatch() bool {
	return m.matchType == termMatchTypeSuffix
}

func (m *TermMatcher) Match(str string) bool {
	l1, l2 := len(m.matchStr), len(str)
	// empty term is always matched.
	if m.IsFullMatch() {
		if l1 != l2 {
			return false
		}
		return m.matchStr == str
	}
	if l1 == 0 {
		return true
	}
	if l2 < l1 {
		return false
	}
	switch {
	case m.IsPrefixMatch():
		return strings.HasPrefix(str, m.matchStr)
	case m.IsSuffixMatch():
		return strings.HasSuffix(str, m.matchStr)
	case m.IsContainMatch():
		return strings.Contains(str, m.matchStr)
	}
	return false
}

func (m *TermMatcher) String() string {
	return m.originalStr
}

func MakeTermMatchers(subs []string) Matchers {
	matchers := make([]Matcher, len(subs))
	for i, str := range subs {
		matchers[i] = NewTermMatcher(str)
	}
	return matchers
}
