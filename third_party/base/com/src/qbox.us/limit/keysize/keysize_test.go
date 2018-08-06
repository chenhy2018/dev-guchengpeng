package keysize

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify.v2/assert"
	"github.com/stretchr/testify.v2/require"
	"qbox.us/limit"
)

func TestKeyLimit(t *testing.T) {

	l := New(20)
	key1 := []byte("1")
	require.Equal(t, uint64(0), l.Running())

	require.NoError(t, l.Acquire(key1, 5))
	require.NoError(t, l.Acquire(key1, 5))
	require.NoError(t, l.Acquire(key1, 5))
	require.Equal(t, limit.ErrLimit, l.Acquire(key1, 10))
	require.NoError(t, l.Acquire(key1, 5))
	require.Equal(t, uint64(20), l.Running())

	l.Release(key1, uint64(5))
	l.Release(key1, uint64(5))
	require.Equal(t, uint64(10), l.Running())
	require.Equal(t, uint64(10), l.AcquireRemains(key1))
	require.Equal(t, uint64(20), l.Running())
	l.Release(key1, uint64(100))
	require.Equal(t, uint64(0), l.Running())
}

func TestKeySizeWriter(t *testing.T) {
	l := New(1024 * 1024 * 20)
	key := []byte("1")
	buf := bytes.NewReader(make([]byte, 1024*1024*100))
	ksr := KeySizeWriter{Key: key, Limit: l, Wrc: nopCloser{ioutil.Discard}}
	_, err := io.Copy(&ksr, buf)
	assert.Equal(t, err, limit.ErrLimit)
	assert.Equal(t, ksr.C <= 1024*1024*20, true)
	l.Release(key, ksr.C)
	assert.Equal(t, l.Running(), uint64(0))

	buf = bytes.NewReader(make([]byte, 1024*1024*10))
	ksr2 := KeySizeWriter{Key: key, Limit: l, Wrc: nopCloser{ioutil.Discard}}
	n, err := io.Copy(&ksr2, buf)
	assert.Equal(t, err, nil)
	assert.Equal(t, n, int64(1024*1024*10))
	assert.Equal(t, ksr2.C, uint64(1024*1024*10))
	assert.Equal(t, l.Running(), uint64(1024*1024*10))
	l.Release(key, ksr2.C)
	assert.Equal(t, l.Running(), uint64(0))
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }
