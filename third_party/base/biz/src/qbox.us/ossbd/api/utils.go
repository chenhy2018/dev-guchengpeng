package api

import (
	"io"
	"math/rand"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gopkg.in/mgo.v2/bson"
)

func newclients(all []*oss.Client) *clients {
	clis := make([]*oss.Client, len(all))
	copy(clis, all)
	return &clients{
		clis:       clis,
		backupClis: clis,
	}
}

type clients struct {
	clis, backupClis []*oss.Client
}

func (p *clients) Get() (client *oss.Client) {
	if len(p.clis) == 0 {
		return p.backupClis[rand.Intn(len(p.backupClis))]
	}
	id := rand.Intn(len(p.clis))
	var p2 []*oss.Client
	p2 = append(p2, p.clis[:id]...)
	p2 = append(p2, p.clis[id+1:]...)
	client, p.clis = p.clis[id], p2
	return
}

func (p *clients) GetBucket(bucketName string) (bucket *oss.Bucket) {
	b, err := p.Get().Bucket(bucketName)
	if err != nil {
		panic(err)
	}
	return b
}

var bufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 32*1024)
		return &b
	},
}

func bcopy(w io.Writer, r io.Reader) (n int64, err error) {
	buf := bufPool.Get().(*[]byte)
	n, err = io.CopyBuffer(w, r, *buf)
	bufPool.Put(buf)
	return
}

func Reverse(s bson.ObjectId) bson.ObjectId {
	r := []byte(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return bson.ObjectId(r)
}
