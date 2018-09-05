package fopd

import (
	"encoding/json"
	"strconv"
	"sync"
	"sync/atomic"
)

type Stats struct {
	items map[string]*Item // key = cmd|host
	mu    sync.Mutex
}

type Item struct {
	Failed  AtomicInt `json:"failed"`
	Retry   AtomicInt `json:"retry"`
	Timeout AtomicInt `json:"timeout"`
}

func NewStats() *Stats {
	return &Stats{items: make(map[string]*Item)}
}

func MakeStatsKey(cmd string, fopMode uint32, host string) string {

	return cmd + "|" + strconv.FormatUint(uint64(fopMode), 10) + "|" + host
}

func (s *Stats) GetItem(key string) *Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := s.items[key]
	if item == nil {
		item = new(Item)
		s.items[key] = item
	}
	return item
}

func (s *Stats) IncFailed(key string) {
	s.GetItem(key).Failed.Add(1)
}

func (s *Stats) IncTimeout(key string) {
	s.GetItem(key).Timeout.Add(1)
}

func (s *Stats) IncRetry(key string) {
	s.GetItem(key).Retry.Add(1)
}

func (s *Stats) Dump() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, _ := json.MarshalIndent(s.items, "", "\t")
	return b
}

func (s *Stats) Clear(keys []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, key := range keys {
		if _, ok := s.items[key]; ok {
			s.items[key] = new(Item)
		}
	}
}

func (s *Stats) ClearAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]*Item)
}

type AtomicInt int64

func (i *AtomicInt) Add(n int64) {
	atomic.AddInt64((*int64)(i), n)
}

func (i *AtomicInt) Get() int64 {
	return atomic.LoadInt64((*int64)(i))
}
