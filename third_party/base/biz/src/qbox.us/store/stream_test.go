package store

import (
	"bytes"
	"crypto/sha1"
	"qbox.us/fh/sha1bd"
	"github.com/qiniu/xlog.v1"
	"testing"
)

func newStream(stg *SimpleStorage, text string, chunkBits uint, t *testing.T) []byte {
	chunkSize := 1 << chunkBits
	n := (len(text) + (chunkSize - 1)) >> chunkBits
	r := bytes.NewBufferString(text)
	keys := bytes.NewBuffer(nil)
	key := sha1.New()
	for i := 1; i < n; i++ {
		key.Reset()
		key.Write([]byte(text)[(i-1)*chunkSize : i*chunkSize])
		err := stg.Put(xlog.NewDummy(), key.Sum(nil), r, chunkSize)
		if err != nil {
			t.Fatal(err)
		}
		keys.Write(key.Sum(nil))
	}
	key.Reset()
	key.Write([]byte(text)[(n-1)*chunkSize : (n-1)*chunkSize+r.Len()])
	err := stg.Put(xlog.NewDummy(), key.Sum(nil), r, r.Len())
	if err != nil {
		t.Fatal(err)
	}
	keys.Write(key.Sum(nil))
	return keys.Bytes()
}

type stream struct {
	g         Getter
	keys      []byte
	chunkBits uint
}

func doRead(p stream, from int64, to int64, v string, t *testing.T) {
	w := bytes.NewBuffer(nil)
	err := sha1bd.StreamRead2(xlog.NewDummy(), p.g, p.keys, p.chunkBits, w, from, to, [4]uint16{0})
	if err != nil {
		t.Fatal(err)
	}
	v2 := string(w.Bytes())
	if v2 != v {
		t.Fatalf("Stream.WriteTo error - data: %v != %v", v2, v)
	}
}

func TestStream(t *testing.T) {
	text := "12345678abcdefg"
	chunkBits := uint(2)
	stg := NewSimpleStorage()
	keys := newStream(stg, text, chunkBits, t)
	p := stream{&SimpleStorage2{stg}, keys, chunkBits}
	doRead(p, 0, 15, "12345678abcdefg", t)
	doRead(p, 1, 3, "23", t)
	doRead(p, 1, 7, "234567", t)
	doRead(p, 1, 8, "2345678", t)
	doRead(p, 1, 14, "2345678abcdef", t)
}
