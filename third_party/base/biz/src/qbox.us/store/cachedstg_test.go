package store

import (
	"os"
	"qbox.us/store/cc"
	"testing"
)

// -----------------------------------------------------------

func TestCachedGetter(t *testing.T) {
	name := "1.cache"
	file, err := os.Create(name)
	if err != nil {
		t.Fatal("Open cache: ", err)
	}
	defer file.Close()
	defer os.Remove(name)
	text := "12345678abcdefg"
	chunkBits := uint(2)
	pool := cc.NewChunkPool(64)
	stg := NewSimpleStorage()
	cached := NewCachedStorage(file, pool, chunkBits, &SimpleStorage2{stg})
	keys := newStream(stg, text, chunkBits, t)
	p := stream{cached, keys, chunkBits}
	doRead(p, 1, 3, "23", t)
	doRead(p, 1, 7, "234567", t)
	doRead(p, 1, 8, "2345678", t)
	doRead(p, 1, 14, "2345678abcdef", t)
	doRead(p, 0, 15, "12345678abcdefg", t)
}

// -----------------------------------------------------------
