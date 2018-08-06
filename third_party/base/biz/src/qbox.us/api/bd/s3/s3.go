package s3

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
	qio "github.com/qiniu/io"

	"github.com/qiniu/xlog.v1"
)

var ErrIoCopyFailed = errors.New("io copy failed")
var ErrServiceUnavailable = errors.New("service unavaliable")

var defaultAttemptTotal = 5 * time.Second
var defaultAttemptDelay = 200 * time.Millisecond

type Config struct {
	ACL       string `json:"acl"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
	Endpoint  string `json:"endpoint"`
}

type S3 struct {
	bucket *s3.Bucket
	acl    s3.ACL
	retrys int
}

func New(cfg *Config, bucket string, retrys int) *S3 {

	auth := aws.Auth{AccessKey: cfg.AccessKey, SecretKey: cfg.SecretKey}
	region := aws.Region{Name: cfg.Region, S3Endpoint: cfg.Endpoint}
	s := s3.New(auth, region)
	acl := s3.ACL(cfg.ACL)
	s3.SetAttemptStrategy(&aws.AttemptStrategy{
		Min:   retrys,
		Total: defaultAttemptTotal,
		Delay: defaultAttemptDelay,
	})
	return &S3{
		bucket: s.Bucket(bucket),
		acl:    acl,
		retrys: retrys,
	}
}

var errX500 = errors.New("500")

func (p *S3) PutEx(xl *xlog.Logger, key []byte, r io.ReaderAt, n int, bds [3]uint16) error {

	var err error
	var xerr error
	defer xl.Xtrack("stg.p", time.Now(), &xerr)

	path := base64.URLEncoding.EncodeToString(key)
	for i := 0; i < p.retrys+1; i++ {
		reader := &qio.Reader{ReaderAt: r}
		err = p.bucket.PutReader(path, reader, int64(n), "", p.acl, s3.Options{})
		if err == nil {
			xerr = nil
			break
		}
		xerr = errX500
		if err1, ok := err.(*s3.Error); ok {
			xerr = errors.New(strconv.Itoa(err1.StatusCode))
		}
		xl.Warnf("S3.PutEx: bucket %v put %v failed failed => %+v", p.bucket.Name, path, err)
		err = ErrServiceUnavailable
	}
	return err
}

func (p *S3) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) error {

	var xerr error
	defer xl.Xtrack("stg.g", time.Now(), &xerr)

	headers := make(http.Header)
	headers.Add("Range", fmt.Sprintf("bytes=%v-%v", from, to-1))

	path := base64.URLEncoding.EncodeToString(key)
	resp, err := p.bucket.GetResponseWithHeaders(path, headers)
	if err != nil {
		xerr = errX500
		xl.Warnf("S3.Get: bucket %v GetResponseWithHeaders failed => %v", p.bucket.Name, err)
		return ErrServiceUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 206 {
		xerr = errors.New(strconv.Itoa(resp.StatusCode))
		xl.Warnf("S3.Get: bucket %v get %v with code %v", p.bucket.Name, path, resp.StatusCode)
		return ErrServiceUnavailable
	}
	n, err := io.Copy(w, resp.Body)
	if err != nil {
		xerr = ErrIoCopyFailed
		xl.Warnf("S3.Get: bucket %v get %v, copy %v bytes and failed => %v", p.bucket.Name, path, n, err)
		return ErrIoCopyFailed
	}
	return nil
}
