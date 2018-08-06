package counter

import (
	. "qbox.us/api/counter/common"
	"time"
)

type CachedClientConfig struct {
	FlushCount    int               `json:"flush_count"`
	FlushMillis   int               `json:"flush_millis"`
	BufferSize    int               `json:"buffer_size"`
	HttpConf      *HttpClientConfig `json:"http"`
	IsFakeCounter bool              `json:"is_fake_counter"`
}

func NewCachedClient(conf *CachedClientConfig) (Counter, error) {

	if conf.IsFakeCounter {
		return NewFakeCounter(), nil
	}

	httpClient, err := NewHttpClient(conf.HttpConf)
	if err != nil {
		return nil, err
	}
	cachedCounter, err := NewCachedCounter(
		httpClient,
		conf.FlushCount,
		time.Duration(conf.FlushMillis)*time.Millisecond,
		conf.BufferSize,
	)
	if err != nil {
		return nil, err
	}

	go flushWorker(cachedCounter)
	return cachedCounter, nil
}
