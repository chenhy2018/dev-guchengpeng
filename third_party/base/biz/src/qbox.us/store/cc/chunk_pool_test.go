package cc

import "testing"

func TestChunkPool(t *testing.T) {
	pool := NewChunkPool(4)
	i1 := pool.Alloc()
	if i1 != 0 {
		t.Fatal("ChunkPool.Alloc error - i1")
	}
	i2 := pool.Alloc()
	if i2 != 1 {
		t.Fatal("ChunkPool.Alloc error - i2")
	}
	pool.Free(i1)
	i3 := pool.Alloc()
	if i3 != 0 {
		t.Fatal("ChunkPool.Alloc error - i3")
	}
	i4 := pool.Alloc()
	if i4 != 2 {
		t.Fatal("ChunkPool.Alloc error - i4")
	}
	i5 := pool.Alloc()
	if i5 != 3 {
		t.Fatal("ChunkPool.Alloc error - i5")
	}
	defer func() {
		if r := recover(); r != nil {
			if r == "ChunkPool.Alloc: out of space" {
				return
			}
		}
		t.Fatal("ChunkPool.Alloc error - outofspace")
	}()
	pool.Alloc()
}
