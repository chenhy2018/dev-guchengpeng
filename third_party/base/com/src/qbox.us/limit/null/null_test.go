package null

import (
	"testing"

	"github.com/stretchr/testify.v1/require"
)

func TestNullLimit(t *testing.T) {
	l := New()
	key := []byte("a")
	for i := 0; i < 1000; i++ {
		err := l.Acquire(key)
		require.NoError(t, err)
	}
	for i := 0; i < 100; i++ {
		l.Release(key)
	}
}
