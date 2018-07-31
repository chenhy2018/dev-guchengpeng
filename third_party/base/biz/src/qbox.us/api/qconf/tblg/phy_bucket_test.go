package tblg

import (
	"testing"

	"github.com/qiniu/rpc.v1"
	"github.com/stretchr/testify.v1/require"
	"github.com/stretchr/testify/assert"

	qconf "qbox.us/qconf/qconfapi"
)

func TestCode631(t *testing.T) {
	old := getPhybuckStub
	getPhybuckStub = func(conn *qconf.Client, l rpc.Logger, ret interface{}, id string, cacheFlags int) (err error) {
		return &rpc.ErrorInfo{Code: 612}
	}
	_, _, err := Client{}.GetPhybuck(nil, 0, "")
	assert.Equal(t, err, ErrNoSuchBucket)
	getPhybuckStub = old
}

func TestBucketInfo(t *testing.T) {
	client := Client{}
	_, err := client.GetBucketInfo(nil, 1000, "/abcd")
	require.Equal(t, err, ErrInvalidBucket)
}
