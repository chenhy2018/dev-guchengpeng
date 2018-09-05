package recordio

import (
	"errors"
	"io"
	"strings"
	"testing"

	"testing/iotest"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
)

func TestErrorRecord(t *testing.T) {
	{
		r := iotest.TimeoutReader(strings.NewReader("ok"))
		r2 := NewErrorRecordReader(r)
		n, err := r2.Read(make([]byte, 1))
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.NoError(t, r2.Error())

		n, err = r2.Read(make([]byte, 1))
		assert.Error(t, err)
		assert.Equal(t, err, r2.Error())
	}
	{
		w := &timeoutWriter{ioutil.Discard, 0}
		w2 := NewErrorRecordWriter(w)
		n, err := w2.Write(make([]byte, 1))
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.NoError(t, w2.Error())

		n, err = w2.Write(make([]byte, 1))
		assert.Error(t, err)
		assert.Equal(t, err, w2.Error())
	}
}

type timeoutWriter struct {
	r     io.Writer
	count int
}

func (r *timeoutWriter) Write(p []byte) (int, error) {
	r.count++
	if r.count == 2 {
		return 0, errors.New("timeout")
	}
	return r.r.Write(p)
}
