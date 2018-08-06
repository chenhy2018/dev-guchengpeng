package ossbd

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"testing"

	"github.com/stretchr/testify.v1/require"
	"gopkg.in/mgo.v2/bson"
)

func TestInstance(t *testing.T) {
	hash := sha1.New()
	hash.Write([]byte("dfsdfds"))
	etag := hash.Sum(nil)
	etag[10] = byte('/')
	fsize := int64(1e8)
	bucket := "bucket_test"
	key := bson.NewObjectId()
	log.Println("key", key)
	instance := NewInstance(fsize, etag, bucket, []byte(string(key)))
	r := require.New(t)
	r.Equal(byte(0x08), instance[0])
	r.Equal(byte(0x96), instance[1])
	r.Equal(fsize, instance.Fsize())
	r.Equal(bucket, instance.Bucket())
	r.Equal(byte(0x96), instance.Etag()[0])
	r.Equal(etag, instance.Etag()[1:])
	r.Equal(key.Hex(), instance.Key())

	s := instance.OssEncode()
	u, err := url.Parse(s)
	r.NoError(err)
	r.Equal("oss", u.Scheme)
	r.Equal(bucket, u.Host)
	r.Equal("/"+key.Hex(), u.Path)
	r.Equal(fmt.Sprint(fsize), u.Query().Get("fsize"))
	r.Equal(base64.URLEncoding.EncodeToString(append([]byte{0x96}, etag...)), u.Query().Get("qetag"))
}
