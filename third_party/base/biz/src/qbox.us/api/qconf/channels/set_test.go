package channels

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sortEqual(t *testing.T, a, b []string, msgAndArgs ...interface{}) bool {
	sort.Strings(a)
	sort.Strings(b)
	return assert.Equal(t, a, b, msgAndArgs)
}

func TestSet(t *testing.T) {
	a := []string{"a1", "a2"}
	b := []string{"hello", "world", "?"}
	c := []string{"hate", "world", "!"}

	sortEqual(t, []string{"a1", "a2"}, union(a, a))
	sortEqual(t, []string{"a1", "a2"}, intersection(a, a))
	sortEqual(t, []string{"a1", "a2", "hello", "world", "?"}, union(a, b))
	sortEqual(t, []string{}, intersection(a, b))

	sortEqual(t, []string{"hello", "hate", "world", "?", "!"}, union(b, c))
	sortEqual(t, []string{"world"}, intersection(b, c))

	sortEqual(t, []string{"a1", "a2", "world"}, union(a, intersection(b, c)))

	sortEqual(t, []string{"hello", "?"}, difference(b, c))
}
