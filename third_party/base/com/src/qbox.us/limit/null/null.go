package null

import "qbox.us/limit"

type nullLimit struct {
	c int32
}

func New() limit.Limit {
	return &nullLimit{}
}

func (l *nullLimit) Running() int {
	return -1
}

func (l *nullLimit) Acquire(key2 []byte) error {
	return nil
}

func (l *nullLimit) Release(key2 []byte) {
}

type nullStringLimit struct {
	c int32
}

func NewStringLimit() limit.StringLimit {
	return &nullStringLimit{}
}

func (l *nullStringLimit) Running() int {
	return -1
}

func (l *nullStringLimit) Acquire(key string) error {
	return nil
}

func (l *nullStringLimit) Release(key string) {
}
