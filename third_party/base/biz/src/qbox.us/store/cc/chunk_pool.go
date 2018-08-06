package cc

import "sync"

// -----------------------------------------------------------

type ChunkPool struct {
	indexs []int32
	free   int32
	mutex  *sync.Mutex
	limit  int32
}

func NewChunkPool(limit int32) *ChunkPool {
	indexs := make([]int32, limit)
	for i := int32(1); i < limit; i++ {
		indexs[i-1] = i
	}
	indexs[limit-1] = -1
	return &ChunkPool{indexs, 0, new(sync.Mutex), limit}
}

func (p *ChunkPool) UseAll() {
	for i := int32(0); i < p.limit; i++ {
		p.indexs[i] = -2
	}
	p.free = -1
}

func (p *ChunkPool) Alloc() int32 {
	p.mutex.Lock()
	t := p.free
	if t < 0 {
		p.mutex.Unlock()
		panic("ChunkPool.Alloc: out of space")
	}
	p.free = p.indexs[t]
	p.mutex.Unlock()
	p.indexs[t] = -2 // in use
	return t
}

func (p *ChunkPool) Free(idx int32) {
	if p.indexs[idx] != -2 {
		panic("ChunkPool.Free: free an unsed chunk")
	}
	p.mutex.Lock()
	p.indexs[idx] = p.free
	p.free = idx
	p.mutex.Unlock()
}
