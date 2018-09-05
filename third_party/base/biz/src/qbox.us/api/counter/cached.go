package counter

import (
	"errors"
	"math"
	"sync/atomic"
	"time"

	"github.com/qiniu/log.v1"
	. "qbox.us/api/counter/common"
)

type CachedCounter struct {
	backend               Counter
	flushCount            int
	flushInterval         time.Duration
	incChannel            chan CounterGroup
	closeCompletedChannel chan bool
}

func NewCachedCounter(backend Counter, flushCount int, flushInterval time.Duration,
	bufferSize int) (*CachedCounter, error) {

	if flushInterval == 0 && flushCount == 0 { // no flushing, very likely to be a bug
		return nil, errors.New("No flushing is specified")
	}

	if flushCount == 0 { // not flush on count
		flushCount = math.MaxInt32
	}

	if flushInterval == 0 { // not flush on time interval
		flushInterval = time.Duration(math.MaxInt64)
	}

	return &CachedCounter{
		backend:               backend,
		flushInterval:         flushInterval,
		flushCount:            flushCount,
		incChannel:            make(chan CounterGroup, bufferSize),
		closeCompletedChannel: make(chan bool),
	}, nil
}

func flushCache(cachedCounter *CachedCounter, cache map[string]CounterMap) {
	for key, counterMap := range cache {
		err := cachedCounter.backend.Inc(key, counterMap)
		if err != nil {
			log.Error("fail to flush counter: ",
				key, counterMap, ";=> ", err)
		}
		atomic.AddInt64(&CounterLag, -(int64(len(counterMap))))
	}
}

func flushWorker(cachedCounter *CachedCounter) {
	cachedCount := 0
	cache := make(map[string]CounterMap)
	timer := time.NewTimer(cachedCounter.flushInterval)
	for {
		select {
		case <-timer.C:
			go flushCache(cachedCounter, cache)
			cache = make(map[string]CounterMap)
			cachedCount = 0
			timer.Reset(cachedCounter.flushInterval)
		case cacheGroup, ok := <-cachedCounter.incChannel:
			if !ok { // channel is closed
				flushCache(cachedCounter, cache)
				cachedCounter.backend.Close()
				cachedCounter.closeCompletedChannel <- true
				return
			}
			counterMap, ok := cache[cacheGroup.Key]
			if !ok {
				counterMap = make(CounterMap)
				cache[cacheGroup.Key] = counterMap
			}
			for key, value := range cacheGroup.Counters {
				counterMap[key] += value
			}
			//log.Info(fmt.Sprintf("Cache<%s>", cacheGroup.Key), cache[cacheGroup.Key])
			cachedCount++
			// now check for flush
			if cachedCount >= cachedCounter.flushCount {
				go flushCache(cachedCounter, cache)
				cache = make(map[string]CounterMap)
				cachedCount = 0
				timer.Reset(cachedCounter.flushInterval)
			}
		}

	}
}

var CounterLag int64

func (counter *CachedCounter) Inc(key string, counters CounterMap) (err error) {
	select {
	case counter.incChannel <- CounterGroup{key, counters}:
		atomic.AddInt64(&CounterLag, int64(len(counters)))
		return nil
	default:
		return errors.New("The inc buffer is full")
	}
}

func (counter *CachedCounter) Get(key string, tags []string) (counters CounterMap, err error) {
	return counter.backend.Get(key, tags)
}

func (counter *CachedCounter) Set(key string, counters CounterMap) (err error) {
	return counter.backend.Set(key, counters)
}

func (counter *CachedCounter) ListPrefix(prefix string, tags []string) (
	counterGroups []CounterGroup, err error) {

	return counter.backend.ListPrefix(prefix, tags)
}

func (counter *CachedCounter) ListRange(start string, end string, tags []string) (
	counterGroups []CounterGroup, err error) {

	return counter.backend.ListRange(start, end, tags)
}

func (counter *CachedCounter) Remove(key string) (err error) {
	return counter.backend.Remove(key)
}

func (counter *CachedCounter) Close() {
	close(counter.incChannel)
	<-counter.closeCompletedChannel
}
