package etag

import (
	"bytes"
	"crypto/sha1"
	"github.com/qiniu/ts"
	"testing"
)

func etagOf(text []byte, chunkSize int) []byte {

	h := sha1.New()
	if len(text) <= chunkSize {
		h.Write(text)
	} else {
		i := 0
		for ; i+chunkSize <= len(text); i += chunkSize {
			t := sha1.New()
			t.Write(text[i : i+chunkSize])
			h.Write(t.Sum(nil))
		}
		if i < len(text) {
			t := sha1.New()
			t.Write(text[i:])
			h.Write(t.Sum(nil))
		}
	}
	return h.Sum(nil)
}

func doTestEtag(text1 string, chunkSize int, tt *testing.T) {

	text := []byte(text1)
	hashv := etagOf(text, chunkSize)

	{
		w := New(chunkSize)
		w.Write(text)
		hashr := w.Sum()
		if !bytes.Equal(hashr, hashv) {
			ts.Fatal(tt, "etag calc error:", text1, chunkSize)
		}
	}
	for n := 1; n < chunkSize*2; n++ {
		w := New(chunkSize)
		t := text
		for {
			if n < len(t) {
				w.Write(t[:n])
				t = t[n:]
			} else {
				w.Write(t)
				break
			}
		}
		hashr := w.Sum()
		if !bytes.Equal(hashr, hashv) {
			ts.Fatal(tt, "etag calc error:", text1, chunkSize, n)
		}
	}
}

func TestEtag(t *testing.T) {

	doTestEtag("", 4, t)
	doTestEtag("abc", 4, t)
	doTestEtag("abcd", 4, t)
	doTestEtag("abcde", 4, t)
	doTestEtag("abcdegegttjfgger", 4, t)
	doTestEtag("abcdegegttjfggeru", 4, t)
	doTestEtag("abcdegegttjfggeruh", 4, t)
	doTestEtag("abcdegegttjfggeruhge", 4, t)
	doTestEtag("abcdegegttjfggeruhgeg", 4, t)
}
