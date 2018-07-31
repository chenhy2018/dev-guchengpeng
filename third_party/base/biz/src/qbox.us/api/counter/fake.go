package counter

import (
	. "qbox.us/api/counter/common"
)

type FakeCounter struct{}

func NewFakeCounter() Counter {
	return &FakeCounter{}
}

func (counter *FakeCounter) Inc(key string, counters CounterMap) (err error) {
	return nil
}

func (counter *FakeCounter) Get(key string, tags []string) (counters CounterMap, err error) {
	return nil, nil
}

func (counter *FakeCounter) Set(key string, counters CounterMap) (err error) {
	return nil
}

func (counter *FakeCounter) ListPrefix(prefix string, tags []string) (
	counterGroups []CounterGroup, err error) {

	return nil, nil
}

func (counter *FakeCounter) ListRange(start string, end string, tags []string) (
	counterGroups []CounterGroup, err error) {

	return nil, nil
}

func (counter *FakeCounter) Remove(key string) (err error) {
	return nil
}

func (counter *FakeCounter) Close() {
	// empty
}
