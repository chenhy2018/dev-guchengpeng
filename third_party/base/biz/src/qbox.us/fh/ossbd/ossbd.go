package ossbd

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"qbox.us/fh/proto"
	"qbox.us/fh/stream"

	xlog "github.com/qiniu/xlog.v1"
)

// Ver(8) + tag(0x96) + Fsize + etag + oss_bucket + / + oss_key

func NewInstance(fsize int64, etag []byte, bucket string, key []byte) (i Instance) {
	if len(etag) != 20 {
		panic("invalid etag")
	}
	if len(key) == 0 {
		panic("invalid key")
	}
	if len(bucket) == 0 {
		panic("invalid bucket")
	}
	if strings.Contains(bucket, "/") {
		panic("bucket cannot contains /")
	}
	i = make([]byte, 1+1+8+20+len(bucket)+1+len(key))
	i[0] = 0x08
	i[1] = 0x96
	binary.LittleEndian.PutUint64(i[2:2+8], uint64(fsize))
	copy(i[10:10+20], etag)
	copy(i[30:30+len(bucket)], []byte(bucket))
	i[30+len(bucket)] = byte('/')
	copy(i[30+len(bucket)+1:], key)
	return
}

type Instance []byte

func (p Instance) Bucket() string {
	index := bytes.IndexByte(p[30:], byte('/'))
	if index < 0 {
		panic("invalid fh")
	}
	return string(p[30 : 30+index])
}

func (p Instance) Key() string {
	return hex.EncodeToString(p.RawKey())
}

func (p Instance) RawKey() []byte {
	index := bytes.IndexByte(p[30:], byte('/'))
	if index < 0 {
		panic("invalid fh")
	}
	return p[30+index+1:]
}

func (p Instance) BdlockerID() string {
	return string(append([]byte(p.Bucket()+"/"), p.RawKey()...))
}

func (p Instance) String() string {
	return p.Base64URLEncode()
}

// oss://bucket/key?fsize=fsize&qetag=qetag
func (p Instance) OssEncode() string {
	return "oss://" + p.Bucket() + "/" + p.Key() + "?" + url.Values{
		"fsize": []string{fmt.Sprintf("%d", p.Fsize())},
		"qetag": []string{base64.URLEncoding.EncodeToString(p.Etag())},
	}.Encode()
}

func (p Instance) Base64StdEncode() string {
	return base64.StdEncoding.EncodeToString(p)
}

func (p Instance) Base64URLEncode() string {
	return base64.URLEncoding.EncodeToString(p)
}

func (p Instance) Ibd() uint16 {

	return 0
}

func (p Instance) Ibdc() uint16 {

	return 0
}

func (p Instance) Etag() []byte {
	etag := make([]byte, 21)
	etag[0] = 0x16
	if p.Fsize() > 1<<22 {
		etag[0] |= 0x80
	}
	copy(etag[1:], p[10:10+20])
	return etag
}

func (p Instance) Fsize() int64 {
	return int64(binary.LittleEndian.Uint64(p[2:10]))
}

var zeroSha1 = sha1.New().Sum(nil)

func (p Instance) Sha1(xl *xlog.Logger, getter *proto.Getter, fsize int64) ([]byte, error) {
	if fsize == 0 {
		return zeroSha1, nil
	}
	if etag := p.Etag(); etag[0]&0x80 == 0 {
		return etag[1:], nil
	}

	r, err := p.Source(xl, getter, fsize)
	if err != nil {
		return nil, err
	}

	h := sha1.New()
	if err = r.RangeRead(h, 0, fsize); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (p Instance) Source(xl *xlog.Logger, getter *proto.Getter, fsize int64) (r proto.Source, err error) {

	fh := []byte(p)
	r = stream.New(xl, getter.Oss, fh, fsize)
	return
}
