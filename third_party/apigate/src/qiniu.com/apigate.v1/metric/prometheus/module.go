package prometheus

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type moduleConfig struct {
	UseCounter bool `json:"use_counter"`

	UseHistogram     bool      `json:"use_histogram"`
	HistogramBuckets []float64 `json:"histogram_buckets"`
}

type module struct {
	counter   *prometheus.CounterVec
	histogram *prometheus.HistogramVec
}

func (m *module) Record(pattern string, code int, costSeconds float64) {

	if m.counter != nil {
		m.counter.With(map[string]string{
			"code":    strconv.Itoa(code),
			"pattern": pattern,
		}).Inc()
	}

	if m.histogram != nil {
		m.histogram.With(map[string]string{
			"code":    strconv.Itoa(code),
			"pattern": pattern,
		}).Observe(costSeconds)
	}
}

// ----------------------------------------------------------------

type requestEvent struct {
	m       *module
	pattern string
	w       http.ResponseWriter

	startAt time.Time
	code    int
}

func (r *requestEvent) AuthParsed(w http.ResponseWriter, req *http.Request) {
}

func (r *requestEvent) End(w http.ResponseWriter, req *http.Request) {

	costSeconds := time.Since(r.startAt).Seconds()
	r.m.Record(r.pattern, r.code, costSeconds)
}

func (r *requestEvent) Header() http.Header {
	return r.w.Header()
}

func (r *requestEvent) Write(b []byte) (int, error) {
	return r.w.Write(b)
}

func (r *requestEvent) WriteHeader(code int) {
	r.code = code
	r.w.WriteHeader(code)
}
