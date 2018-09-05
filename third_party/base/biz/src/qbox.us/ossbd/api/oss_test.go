package api

import (
	"bytes"
	"crypto/md5"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	xlog "github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify.v1/require"
	"gopkg.in/mgo.v2/bson"
	"qbox.us/fh/ossbd"
)

func TestOss(t *testing.T) {
	r := require.New(t)
	cli, err := New(&Config{
		Endpoint:          "http://oss-cn-hangzhou.aliyuncs.com",
		Proxies:           []string{"http://oss-cn-hangzhou.aliyuncs.com", "http://127.0.0.1:123", "http://127.0.0.1:345"},
		AK:                "LTAI1iMaRih3tNpY",
		SK:                "wqHgvmmTcuzW6iBUywcJAT1A3TBFFE",
		Bucket:            "kodo-test",
		ConnectTimeoutS:   30,
		ReadWriteTimeoutS: 30,
		RetryTimes:        3,
	})
	r.NoError(err)
	xl := xlog.NewDummy()
	// put
	data := "testdsflsdhfsdfhdsfskdhfsdkfsdkfsd"
	expectMd5 := md5.Sum([]byte(data))
	fh, md5, err := cli.Put(xl, strings.NewReader(data), int64(len(data)))
	r.NoError(err)
	require.Equal(t, expectMd5[:], md5)
	defer cli.Delete(xl, fh)
	xl.Info(ossbd.Instance(fh))
	// get
	rc, fsize, err := cli.Get(xl, fh, 0, int64(len(data)))
	r.NoError(err)
	defer rc.Close()
	r.Equal(int(fsize), len(data))
	data2, err := ioutil.ReadAll(rc)
	r.NoError(err)
	r.Equal(data, string(data2))
	// get range 0-0
	rc, fsize, err = cli.Get(xl, fh, 0, 1)
	r.NoError(err)
	defer rc.Close()
	r.Equal(1, int(fsize))
	data2, err = ioutil.ReadAll(rc)
	r.NoError(err)
	r.Equal(data[0:1], string(data2))
	// get range 0-4
	rc, fsize, err = cli.Get(xl, fh, 0, 3)
	r.NoError(err)
	defer rc.Close()
	r.Equal(3, int(fsize))
	data2, err = ioutil.ReadAll(rc)
	r.NoError(err)
	r.Equal(data[0:3], string(data2))
	// delete
	err = cli.Delete(xl, fh)
	r.NoError(err)
	// after delete, get return 404
	_, _, err = cli.Get(xl, fh, 0, int64(len(data)))
	r.Equal(ErrDocumentNotFound, err)
	// delete again
	err = cli.Delete(xl, fh)
	r.NoError(err)
}

func TestBackup(t *testing.T) {
	r := require.New(t)
	cli, err := New(&Config{
		Endpoint:          "http://oss-cn-hangzhou.aliyuncs.com",
		Proxies:           []string{"http://127.0.0.1:123", "http://127.0.0.1:345"},
		BackupEndpoint:    "https://oss-cn-hangzhou.aliyuncs.com",
		RetryPutBackup:    true,
		AK:                "LTAI1iMaRih3tNpY",
		SK:                "wqHgvmmTcuzW6iBUywcJAT1A3TBFFE",
		Bucket:            "kodo-test",
		ConnectTimeoutS:   30,
		ReadWriteTimeoutS: 30,
		RetryTimes:        3,
	})
	r.NoError(err)
	xl := xlog.NewDummy()
	// put
	data := "testdsflsdhfsdfhdsfskdhfsdkfsdkfsd"
	expectMd5 := md5.Sum([]byte(data))
	fh, md5, err := cli.Put(xl, strings.NewReader(data), int64(len(data)))
	require.Equal(t, expectMd5[:], md5)
	r.NoError(err)
	defer cli.Delete(xl, fh)
	xl.Info(ossbd.Instance(fh))
	// get
	rc, fsize, err := cli.Get(xl, fh, 0, int64(len(data)))
	r.NoError(err)
	defer rc.Close()
	r.Equal(int(fsize), len(data))
	data2, err := ioutil.ReadAll(rc)
	r.NoError(err)
	r.Equal(data, string(data2))
	// get md5
	rc, md5, fsize, err = cli.GetWithMd5(xl, fh, 0, int64(len(data)))
	r.NoError(err)
	defer rc.Close()
	r.Equal(int(fsize), len(data))
	data2, err = ioutil.ReadAll(rc)
	r.NoError(err)
	r.Equal(data, string(data2))
	require.Equal(t, expectMd5[:], md5)
	// get range 0-0
	rc, fsize, err = cli.Get(xl, fh, 0, 1)
	r.NoError(err)
	defer rc.Close()
	r.Equal(1, int(fsize))
	data2, err = ioutil.ReadAll(rc)
	r.NoError(err)
	r.Equal(data[0:1], string(data2))
	// get range 0-3
	rc, fsize, err = cli.Get(xl, fh, 0, 3)
	r.NoError(err)
	defer rc.Close()
	r.Equal(3, int(fsize))
	data2, err = ioutil.ReadAll(rc)
	r.NoError(err)
	r.Equal(data[0:3], string(data2))
	// delete
	err = cli.Delete(xl, fh)
	r.NoError(err)
	// after delete, get return 404
	_, _, err = cli.Get(xl, fh, 0, int64(len(data)))
	r.Equal(ErrDocumentNotFound, err)
	// delete again
	err = cli.Delete(xl, fh)
	r.NoError(err)
}

func TestTimeOut(t *testing.T) {

	ver := runtime.Version()
	if ver >= "go1.10" && ver < "go1.11" {
		t.Log("TestTimeOut: testcase can't pass on go1.10.x; To be followed.")
		return
	}

	xl := xlog.NewDummy()
	r := require.New(t)
	ln, err := net.Listen("tcp", ":0")
	r.NoError(err)
	defer ln.Close()
	addr := ln.Addr().String()
	cli, err := New(&Config{
		Endpoint:          "http://oss-cn-hangzhou.aliyuncs.com",
		Proxies:           []string{"http://" + addr},
		BackupEndpoint:    "https://oss-cn-hangzhou.aliyuncs.com",
		RetryPutBackup:    true,
		AK:                "LTAI1iMaRih3tNpY",
		SK:                "wqHgvmmTcuzW6iBUywcJAT1A3TBFFE",
		Bucket:            "kodo-test",
		ConnectTimeoutS:   1,
		ReadWriteTimeoutS: 1,
		RetryTimes:        0,
	})
	r.NoError(err)
	now := time.Now()
	data := "testdsflsdhfsdfhdsfskdhfsdkfsdkfsd"
	fh, _, err := cli.Put(xl, strings.NewReader(data), int64(len(data)))
	r.NoError(err)
	defer func() { go cli.Delete(xl, fh) }()
	xl.Info(ossbd.Instance(fh))
	r.True(time.Since(now) > time.Second)
	r.True(time.Since(now) < 2*time.Second)
	err = cli.Delete(xl, fh)
	r.NoError(err)
}

func TestPutBigFile(t *testing.T) {
	r := require.New(t)
	cli, err := New(&Config{
		Endpoint:          "http://oss-cn-hangzhou.aliyuncs.com",
		Proxies:           []string{"http://oss-cn-hangzhou.aliyuncs.com", "http://127.0.0.1:123", "http://127.0.0.1:345"},
		AK:                "LTAI1iMaRih3tNpY",
		SK:                "wqHgvmmTcuzW6iBUywcJAT1A3TBFFE",
		Bucket:            "kodo-test",
		ConnectTimeoutS:   30,
		ReadWriteTimeoutS: 30,
		RetryTimes:        3,
		PutSplitSize:      101 * 1024,
	})
	r.NoError(err)
	xl := xlog.NewDummy()
	// put
	for _, size := range []int{101 * 1024, 102 * 1024} {
		data := make([]byte, size)
		_, err = rand.Read(data)
		r.NoError(err)
		expectMd5 := md5.Sum(data)
		log.Printf("%X", expectMd5)
		fh, md5, err := cli.putBigFile(xl, bytes.NewReader(data), int64(len(data)))
		require.Equal(t, expectMd5[:], md5)
		r.NoError(err)
		defer cli.Delete(xl, fh)
		xl.Info(ossbd.Instance(fh))
		// get
		rc, fsize, err := cli.Get(xl, fh, 0, int64(len(data)))
		r.NoError(err)
		defer rc.Close()
		r.Equal(int(fsize), len(data))
		data2, err := ioutil.ReadAll(rc)
		r.NoError(err)
		r.Equal(data, data2)
		// get md5
		rc, md5, fsize, err = cli.GetWithMd5(xl, fh, 0, int64(len(data)))
		r.NoError(err)
		defer rc.Close()
		r.Equal(int(fsize), len(data))
		data2, err = ioutil.ReadAll(rc)
		r.NoError(err)
		r.Equal(data, data2)
		r.Nil(md5)
		// get range 0-1
		rc, fsize, err = cli.Get(xl, fh, 0, 1)
		r.NoError(err)
		defer rc.Close()
		r.Equal(1, int(fsize))
		data2, err = ioutil.ReadAll(rc)
		r.NoError(err)
		r.Equal(data[0:1], data2)
		// get range 0-3
		rc, fsize, err = cli.Get(xl, fh, 0, 3)
		r.NoError(err)
		defer rc.Close()
		r.Equal(3, int(fsize))
		data2, err = ioutil.ReadAll(rc)
		r.NoError(err)
		r.Equal(data[0:3], data2)
		// delete
		err = cli.Delete(xl, fh)
		r.NoError(err)
		// after delete, get return 404
		_, _, err = cli.Get(xl, fh, 0, int64(len(data)))
		r.Equal(ErrDocumentNotFound, err)
		// delete again
		err = cli.Delete(xl, fh)
		r.NoError(err)
	}
	data := make([]byte, 102*1024)
	_, err = rand.Read(data)
	r.NoError(err)
	cli, err = New(&Config{
		Endpoint:          "http://oss-cn-hangzhou.aliyuncs.com",
		Proxies:           []string{"http://oss-cn-hangzhou.aliyuncs.com", "http://127.0.0.1:123", "http://127.0.0.1:345"},
		AK:                "LTAI1iMaRih3tNpY",
		SK:                "wqHgvmmTcuzW6iBUywcJAT1A3TBFFE",
		Bucket:            "kodo-test",
		ConnectTimeoutS:   30,
		ReadWriteTimeoutS: 1,
		RetryTimes:        3,
		PutSplitSize:      101 * 1024,
	})
	r.NoError(err)
	var imur oss.InitiateMultipartUploadResult
	testfunc = func(initRet oss.InitiateMultipartUploadResult) {
		imur = initRet
	}
	fh, _, err := cli.putBigFile(xl, bytes.NewReader(data), int64(len(data)+1))
	r.Equal(ErrPutFailed, err)
	r.Equal([]byte(nil), fh)
	bucket, err := cli.clis[0].Bucket(cli.conf.Bucket)
	r.NoError(err)
	_, err = bucket.ListUploadedParts(imur)
	r.Error(err)
	r.True(strings.Contains(err.Error(), "ErrorCode=NoSuchUpload"), err.Error())
}

func TestReverse(t *testing.T) {
	{
		s := bson.ObjectId("abcdef")
		s1 := Reverse(s)
		require.Equal(t, bson.ObjectId("fedcba"), s1)
		s2 := Reverse(s1)
		require.Equal(t, s, s2)
	}

	{
		s := bson.ObjectId("abcde")
		s1 := Reverse(s)
		require.Equal(t, bson.ObjectId("edcba"), s1)
		s2 := Reverse(s1)
		require.Equal(t, s, s2)
	}
}
