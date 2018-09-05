package keycount

import (
	"sync/atomic"
	"testing"
	"time"

	"qbox.us/limit"

	"github.com/stretchr/testify.v1/require"
)

func TestKeyLimit(t *testing.T) {

	l := New(2)
	key1, key2, key3 := []byte("1"), []byte("2"), []byte("3")
	require.Equal(t, 0, l.Running())

	require.NoError(t, l.Acquire(key1))
	require.NoError(t, l.Acquire(key2))
	require.NoError(t, l.Acquire(key1))
	require.NoError(t, l.Acquire(key2))
	require.NoError(t, l.Acquire(key3))
	require.Equal(t, limit.ErrLimit, l.Acquire(key2))

	require.Equal(t, 5, l.Running())

	require.Equal(t, limit.ErrLimit, l.Acquire(key1))
	require.Equal(t, limit.ErrLimit, l.Acquire(key2))
	require.NoError(t, l.Acquire(key3))
	require.Equal(t, 6, l.Running())

	l.Release(key1)
	l.Release(key2)
	require.Equal(t, 4, l.Running())
	require.NoError(t, l.Acquire(key1))
	require.NoError(t, l.Acquire(key2))
	require.Equal(t, limit.ErrLimit, l.Acquire(key3))
	require.Equal(t, 6, l.Running())
}

func TestBlockingKeyLimit(t *testing.T) {

	l := NewBlockingKeyCountLimit(2)
	key1, key2, key3 := []byte("1"), []byte("2"), []byte("3")
	done1, done2, done3 := atomicBool{}, atomicBool{}, atomicBool{}
	require.Equal(t, 0, l.Running())

	require.NoError(t, l.Acquire(key1))
	require.NoError(t, l.Acquire(key2))
	require.NoError(t, l.Acquire(key1))
	require.NoError(t, l.Acquire(key2))
	require.NoError(t, l.Acquire(key3))
	go func() {
		require.NoError(t, l.Acquire(key2)) // blocking
		done2.Set(true)
	}()

	time.Sleep(0.5e9)
	require.Equal(t, 5, l.Running())

	go func() {
		require.NoError(t, l.Acquire(key1)) // blocking
		done1.Set(true)
	}()
	time.Sleep(0.5e9)
	require.NoError(t, l.Acquire(key3))
	require.Equal(t, 6, l.Running())
	require.False(t, done1.Get())
	require.False(t, done2.Get())

	l.Release(key1)
	l.Release(key2)
	time.Sleep(0.5e9)
	require.True(t, done1.Get())
	require.True(t, done2.Get())
	require.Equal(t, 6, l.Running())

	done1, done2, done3 = atomicBool{}, atomicBool{}, atomicBool{}
	go func() {
		require.NoError(t, l.Acquire(key1)) // blocking
		done1.Set(true)
	}()
	go func() {
		require.NoError(t, l.Acquire(key2)) // blocking
		done2.Set(true)
	}()
	go func() {
		require.NoError(t, l.Acquire(key3)) // blocking
		done3.Set(true)
	}()
	time.Sleep(0.5e9)
	require.False(t, done1.Get())
	require.False(t, done2.Get())
	require.False(t, done3.Get())
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

func TestSyncKeyLimit(t *testing.T) {

	l := NewSync(2)
	key1, key2, key3 := string("1"), string("2"), string("3")
	require.Equal(t, 0, l.Running())

	require.NoError(t, l.Acquire(key1))
	require.NoError(t, l.Acquire(key2))
	require.NoError(t, l.Acquire(key1))
	require.NoError(t, l.Acquire(key2))
	require.NoError(t, l.Acquire(key3))
	require.Equal(t, limit.ErrLimit, l.Acquire(key2))

	require.Equal(t, 5, l.Running())

	require.Equal(t, limit.ErrLimit, l.Acquire(key1))
	require.Equal(t, limit.ErrLimit, l.Acquire(key2))
	require.NoError(t, l.Acquire(key3))
	require.Equal(t, 6, l.Running())

	l.Release(key1)
	l.Release(key2)
	require.Equal(t, 4, l.Running())
	require.NoError(t, l.Acquire(key1))
	require.NoError(t, l.Acquire(key2))
	require.Equal(t, limit.ErrLimit, l.Acquire(key3))
	require.Equal(t, 6, l.Running())
}
