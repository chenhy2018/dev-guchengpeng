package match_test

import (
	"testing"

	"github.com/stretchr/testify.v2/assert"
	"qbox.us/iam/decider/match"
)

func TestTermMatcher(t *testing.T) {
	var testCases = []struct {
		partStr     string
		matchStr    string
		matchExpect bool
		matchTypeFn func(*match.TermMatcher) bool
	}{
		{"", "abcd", true, (*match.TermMatcher).IsContainMatch},
		{"*", "abcd", true, (*match.TermMatcher).IsContainMatch},
		{"ab*", "abcd", true, (*match.TermMatcher).IsPrefixMatch},
		{"*cd", "abcd", true, (*match.TermMatcher).IsSuffixMatch},
		{"*ab", "abcd", false, (*match.TermMatcher).IsSuffixMatch},
		{"*bc*", "abcd", true, (*match.TermMatcher).IsContainMatch},
		{"ab", "abcd", false, (*match.TermMatcher).IsFullMatch},
		{"ab", "ac", false, (*match.TermMatcher).IsFullMatch},
		{"ab", "", false, (*match.TermMatcher).IsFullMatch},
		{"a*b", "acb", false, (*match.TermMatcher).IsFullMatch},
		{"*abcdefg", "abcd", false, (*match.TermMatcher).IsSuffixMatch},
		{"ab", "*", false, (*match.TermMatcher).IsFullMatch},
		{"*ab", "*b", false, (*match.TermMatcher).IsSuffixMatch},
		{"ab*", "a*", false, (*match.TermMatcher).IsPrefixMatch},
		{"a*b", "a*b", true, (*match.TermMatcher).IsFullMatch},
	}
	for _, testCase := range testCases {
		matcher := match.NewTermMatcher(testCase.partStr)
		assert.True(t, testCase.matchTypeFn(matcher), "testCase: %+v", testCase)
		assert.Equal(t, testCase.partStr, matcher.String(), "testCase: %+v", testCase)
		assert.Equal(t, testCase.matchExpect, matcher.Match(testCase.matchStr), "testCase: %+v", testCase)
	}
}

func TestTermMatchers(t *testing.T) {
	var testCases = []struct {
		partStrs      []string
		matchStrs     []string
		expectMatched bool
	}{
		{[]string{"A", "B", "C", "D"}, []string{"A", "B", "C", "D"}, true},
		{[]string{"A", "B", "C", "D"}, []string{"A", "B", "C"}, false},
		{[]string{"A", "B", "C", "D"}, []string{"A", "B", "C", "E"}, false},
		{[]string{"A"}, []string{"A"}, true},
		{[]string{"A"}, []string{"B"}, false},
		{[]string{}, []string{}, true},
	}
	for _, testCase := range testCases {
		matchers := match.MakeTermMatchers(testCase.partStrs)
		matched := matchers.MatchAll(testCase.matchStrs)
		assert.Equal(t, testCase.expectMatched, matched)
	}
}
