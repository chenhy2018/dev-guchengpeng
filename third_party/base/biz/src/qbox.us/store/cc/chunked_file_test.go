package cc

import (
	"os"
	"testing"
)

func TestChunkedFile(t *testing.T) {
	name := "1.chunked"
	file, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	defer os.Remove(name)
	pool := NewChunkPool(16)
	w := NewChunkedFile(file, pool, 3, 2)
	ops := []string{
		"123",
		"4567",
		"89012",
		"34",
	}
	for _, op := range ops {
		n, err := w.Write([]byte(op))
		if n != len(op) || err != nil {
			t.Fatal("ChunkedFile.Write failed")
		}
	}
	b := w.Buf[:2]
	if string(b) != "34" {
		t.Fatal("ChunkedFile.Buffer failed: " + string(b))
	}
	chunkNo, chunkSize := w.Chunk()
	if chunkNo != 1 || chunkSize != 6 {
		t.Fatal("ChunkedFile.Chunk failed")
	}
	w.Sync()
}
