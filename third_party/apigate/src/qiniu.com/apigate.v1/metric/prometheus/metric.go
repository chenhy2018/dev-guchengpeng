package prometheus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	. "github.com/qiniu/apigate.v1/proto"
	"qbox.us/localProxy"
)

type Config struct {
	PushIntervalS   int  `json:"push_interval_s"`
	InstanceUseHost bool `json:"instance_use_host"`

	Gateway struct {
		Job       string `json:"job"`
		Instance  string `json:"instance"`
		URL       string `json:"url"`
		Proxy     string `json:"proxy"`
		pushEntry string `json:"-"`

		LocalProxyAddr string `json:"localProxyAddr"`
	} `json:"gateway"`
}

type Prometheus struct {
	Config
	registry prometheus.Registry
	closed   chan struct{}

	metrics map[string]*module
	lock    sync.RWMutex
}

func New(conf Config) (p *Prometheus, err error) {

	if conf.InstanceUseHost {
		conf.Gateway.Instance, err = os.Hostname()
		if err != nil {
			return
		}
	}

	conf.Gateway.pushEntry = conf.Gateway.URL
	if conf.Gateway.Proxy != "" {
		if conf.Gateway.LocalProxyAddr == "" {
			conf.Gateway.LocalProxyAddr = "127.0.0.1:2782"
		}
		localProxyConfig := localProxy.Config{
			Addr:      conf.Gateway.URL,
			ProxyAddr: conf.Gateway.Proxy,
			LocalAddr: conf.Gateway.LocalProxyAddr,
		}
		err := localProxy.Listen(localProxyConfig)
		if err != nil {
			return nil, err
		}
		log.Printf("prometheus proxy run @%s", conf.Gateway.LocalProxyAddr)
		conf.Gateway.pushEntry = conf.Gateway.LocalProxyAddr
	}

	return &Prometheus{
		Config: conf,
		registry: prometheus.NewRegistry(prometheus.NewRegistryOption{
			BufPool:          make(chan *bytes.Buffer, 4),
			MetricFamilyPool: make(chan *dto.MetricFamily, 1000),
			MetricPool:       make(chan *dto.Metric, 10000),
		}),
		closed:  make(chan struct{}, 1),
		metrics: make(map[string]*module),
	}, nil
}

func (p *Prometheus) Run() {

	for range time.Tick(time.Second * time.Duration(p.PushIntervalS)) {
		select {
		case <-p.closed:
			log.Println("stop push")
			return
		default:
		}

		err := p.registry.Push(p.Gateway.Job, p.Gateway.Instance, p.Gateway.pushEntry, "PUT")
		if err != nil {
			log.Println("push failed:", err)
		}
	}
}

func (p *Prometheus) Close() {

	log.Println("try close")
	close(p.closed)
}

func (p *Prometheus) OpenRequest(ctx context.Context, w *http.ResponseWriter, req *http.Request) RequestEvent {

	mod := ModFromContextSafe(ctx)
	p.lock.RLock()
	defer p.lock.RUnlock()
	if m, ok := p.metrics[mod]; ok {
		pattern := PatternFromContextSafe(ctx)
		nw := &requestEvent{
			m:       m,
			pattern: pattern,
			w:       *w,
			startAt: time.Now(),
			code:    200,
		}
		*w = nw
		return nw
	}

	return NilReqEvent
}

func (p *Prometheus) unregister(name string) {

	p.lock.RLock()
	defer p.lock.RUnlock()

	if preMetric, ok := p.metrics[name]; ok {
		if preMetric.counter != nil {
			p.registry.Unregister(preMetric.counter)
		}
		if preMetric.histogram != nil {
			p.registry.Unregister(preMetric.histogram)
		}
	}
}

func (p *Prometheus) Register(name string, confStr string) error {

	if confStr == "" {
		log.Printf("module(%s) has no metric conf", name)
		return nil
	}
	var conf moduleConfig
	err := json.Unmarshal([]byte(confStr), &conf)
	if err != nil {
		return err
	}

	p.unregister(name)

	// conf from string
	m := &module{}

	if conf.UseCounter {
		m.counter = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: metricName(counterNamePrefix, name),
			Help: fmt.Sprintf("%s's API counter", name),
		}, []string{"code", "pattern"})

		p.registry.MustRegister(m.counter)
	}

	if conf.UseHistogram {
		m.histogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    metricName(histogramNamePrefix, name),
			Help:    fmt.Sprintf("histogram of %s's API time cost", name),
			Buckets: conf.HistogramBuckets,
		}, []string{"code", "pattern"})

		p.registry.MustRegister(m.histogram)
	}

	p.lock.Lock()
	if !conf.UseCounter && !conf.UseHistogram {
		delete(p.metrics, name)
	} else {
		p.metrics[name] = m
	}
	p.lock.Unlock()
	return nil
}

// ----------------------------------------------------------------

const (
	counterNamePrefix   = "apigate_requests_count_"
	histogramNamePrefix = "apigate_request_duration_sec_hisogram_"
)

func metricName(prefix, moduleName string) string {

	mn := strings.Replace(moduleName, "-", "", -1)
	mn = strings.Replace(mn, "_", "", -1)

	name := prefix + mn
	name = strings.ToLower(name)
	if ok, _ := regexp.MatchString("[a-z0-9]+", name); !ok {
		panic("invalid moduleName: " + moduleName)
	}

	return name
}
