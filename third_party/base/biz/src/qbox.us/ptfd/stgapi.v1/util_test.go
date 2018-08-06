package stgapi

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyExcept(t *testing.T) {

	ss := []string{"a", "b", "c", "d", "e"}
	idxs := []int{0, 1, 2, 3, 4}
	rets := [][]string{
		{"b", "c", "d", "e"},
		{"a", "c", "d", "e"},
		{"a", "b", "d", "e"},
		{"a", "b", "c", "e"},
		{"a", "b", "c", "d"},
	}
	for i, idx := range idxs {
		ret := copyExcept(ss, idx)
		assert.Equal(t, rets[i], ret, "%v", i)
	}
}

func TestRandomShrink(t *testing.T) {

	all := []string{"a", "b", "c", "d", "e"}
	ss := []string{"a", "b", "c", "d", "e"}
	s := ""
	all0 := []string{}
	for _ = range all {
		ss, s = randomShrink(ss)
		all0 = append(all0, s)
	}
	sort.StringSlice(all0).Sort()
	assert.Equal(t, all, all0)
}
