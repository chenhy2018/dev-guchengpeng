package s3

import (
	"bytes"
	"testing"
	"time"

	"github.com/crowdmob/goamz/s3"
	"github.com/crowdmob/goamz/s3/s3test"
	"github.com/stretchr/testify/assert"

	"github.com/qiniu/xlog.v1"
)

func TestS3(t *testing.T) {

	defaultAttemptTotal = 300 * time.Millisecond
	defaultAttemptDelay = 100 * time.Millisecond

	srv, _ := s3test.NewServer(&s3test.Config{})

	cfg := &Config{
		ACL:       "private",
		AccessKey: "abc",
		SecretKey: "efg",
		Region:    "faux-region-1",
		Endpoint:  srv.URL(),
	}
	s := New(cfg, "test", 1)

	var bds3 [3]uint16
	var bds4 [4]uint16
	xl := xlog.NewDummy()
	r := bytes.NewReader([]byte("hello world"))
	err := s.PutEx(xl, []byte("key1"), r, r.Len(), bds3)
	assert.Equal(t, err, ErrServiceUnavailable)

	w := bytes.NewBuffer(nil)
	err = s.Get(xl, []byte("key1"), w, 0, r.Len(), bds4)
	assert.Equal(t, err, ErrServiceUnavailable)

	s.bucket.S3LocationConstraint = true
	err = s.bucket.PutBucket(s3.Private)
	assert.NoError(t, err)

	w = bytes.NewBuffer(nil)
	err = s.Get(xl, []byte("key1"), w, 0, w.Len(), bds4)
	assert.Equal(t, err, ErrServiceUnavailable)

	r = bytes.NewReader([]byte("hello world"))
	err = s.PutEx(xl, []byte("key1"), r, r.Len(), bds3)
	assert.NoError(t, err)

	w = bytes.NewBuffer(nil)
	err = s.Get(xl, []byte("key1"), w, 0, 11, bds4)
	assert.NoError(t, err)
	assert.Equal(t, w.Bytes(), []byte("hello world"))

	w = bytes.NewBuffer(nil)
	err = s.Get(xl, []byte("key1"), w, 3, 8, bds4)
	assert.NoError(t, err)
	assert.Equal(t, w.Bytes(), []byte("lo wo"))
}
