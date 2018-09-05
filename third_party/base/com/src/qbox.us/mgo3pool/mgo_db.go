package mgo3pool

import (
	"sync/atomic"

	"labix.org/v2/mgo"
	"qbox.us/mgo3"
)

type Config mgo3.Config

type Session struct {
	*mgo.Session
	pool  []*mgo3.Session
	n     uint32
	index uint32
}

func Open(cfg *Config, poolSize int) *Session {
	if poolSize < 1 {
		panic("poolSize < 1")
	}
	cfg2 := mgo3.Config(*cfg)
	s := mgo3.Open(&cfg2)
	pool := make([]*mgo3.Session, poolSize)
	pool[0] = s
	for i := 1; i < poolSize; i++ {
		c2 := s.Coll.Copy()
		s2 := &mgo3.Session{Session: c2.Database.Session, DB: c2.Database, Coll: c2}
		pool[i] = s2
	}
	return &Session{Session: s.Session, pool: pool, n: uint32(poolSize)}
}

func (s *Session) Coll() *mgo3.Collection {
	index := atomic.AddUint32(&s.index, 1) % s.n
	return &s.pool[index].Coll
}

func (s *Session) Close() {
	for _, s2 := range s.pool {
		s2.Close()
	}
}
