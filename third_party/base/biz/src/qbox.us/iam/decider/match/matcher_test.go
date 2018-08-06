package match_test

import (
	"testing"

	"github.com/stretchr/testify.v2/assert"
	"qbox.us/iam/decider/match"
)

type mockMatcher string

func (m mockMatcher) Match(str string) bool {
	return string(m) == str
}

func TestMatchers(t *testing.T) {
	strs := []string{"A", "B", "C", "D"}
	matchers := match.Matchers{}
	for _, str := range strs {
		matchers = append(matchers, mockMatcher(str))
	}
	assert.True(t, matchers.MatchAll(strs), "MatchAll(%+v)", strs)
	assert.False(t, matchers.MatchAll(strs[1:]), "MatchAll(%+v)", strs[1:])
	assert.Equal(t, len(matchers), matchers.Len(), "Len()")
	{
		breaked := matchers.Each(func(i int, matcher match.Matcher) bool {
			matched := matcher.Match(strs[i])
			assert.True(t, matched, "Each->Match(%s)", strs[i])
			return false
		})
		assert.False(t, breaked)
	}
	{
		n := len(strs) / 2
		breaked := matchers.Each(func(i int, matcher match.Matcher) bool {
			n--
			matched := matcher.Match(strs[i])
			assert.True(t, matched, "Each->Match(%s)", strs[i])
			return n == 0
		})
		assert.True(t, breaked)
		assert.Zero(t, n)
	}
}
