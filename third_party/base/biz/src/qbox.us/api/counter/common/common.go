package backend

type CounterMap map[string]int64

type CounterGroup struct {
	Key      string     `json:"key"`
	Counters CounterMap `json:"counters"`
}

type Counter interface {
	Inc(key string, counters CounterMap) (err error)
	Get(key string, tags []string) (counters CounterMap, err error)
	Set(key string, counters CounterMap) (err error)
	ListPrefix(prefix string, tags []string) (counterGroups []CounterGroup, err error)
	ListRange(start string, end string, tags []string) (counterGroups []CounterGroup, err error)
	Remove(key string) (err error)
	Close()
}
