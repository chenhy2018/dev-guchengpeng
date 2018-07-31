package count

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify.v1/require"
)

func init() {
	runtime.GOMAXPROCS(4)
}

func TestCountLimit(t *testing.T) {

	l := New(2)
	key := []byte("a")
	require.Equal(t, 0, l.Running())

	require.NoError(t, l.Acquire(key))
	require.NoError(t, l.Acquire(key))
	require.Error(t, l.Acquire(key))
	require.Equal(t, 2, l.Running())

	l.Release(key)
	require.Equal(t, 1, l.Running())
	require.NoError(t, l.Acquire(key))
	require.Error(t, l.Acquire(key))
	require.Equal(t, 2, l.Running())

	l.Release(key)
	l.Release(key)
	require.Equal(t, 0, l.Running())
	require.NoError(t, l.Acquire(key))
	require.NoError(t, l.Acquire(key))
	require.Error(t, l.Acquire(key))
	require.Equal(t, 2, l.Running())
}

func TestBlockingCountLimit(t *testing.T) {
	l := NewBlockingCount(2)
	key := []byte("a")
	require.Equal(t, 0, l.Running())

	require.NoError(t, l.Acquire(key))
	require.NoError(t, l.Acquire(key))
	require.Equal(t, 2, l.Running())

	l.Release(key)
	require.Equal(t, 1, l.Running())
	require.NoError(t, l.Acquire(key))
	require.Equal(t, 2, l.Running())

	l.Release(key)
	l.Release(key)
	require.Equal(t, 0, l.Running())
	require.NoError(t, l.Acquire(key))
	require.NoError(t, l.Acquire(key))
	require.Equal(t, 2, l.Running())

	done := &atomicBool{}
	go func() {
		require.NoError(t, l.Acquire(key)) // blocking
		done.Set(true)
	}()
	time.Sleep(.5e9)
	require.Equal(t, 2, l.Running())
	require.False(t, done.Get())
	l.Release(key)
	time.Sleep(.5e9)
	require.Equal(t, 2, l.Running())
	require.True(t, done.Get())
}

type atomicBool struct {
	ret int32
}

func (self *atomicBool) Set(v bool) {
	if v {
		atomic.StoreInt32(&self.ret, 1)
	} else {
		atomic.StoreInt32(&self.ret, 0)
	}
}

func (self *atomicBool) Get() (v bool) {
	return atomic.LoadInt32(&self.ret) == 1
}
