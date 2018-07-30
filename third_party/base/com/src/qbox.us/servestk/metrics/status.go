package metrics

import (
	"encoding/json"
	"net/http"
	"time"

	"qbox.us/servestk"

	"github.com/qiniu/http/httputil.v1"
	"github.com/rcrowley/go-metrics"
)

type Status struct {
	Counter metrics.Counter
	Timer   metrics.Timer
}

func NewStatus() *Status {
	return &Status{
		Counter: metrics.NewRegisteredCounter("processing", nil),
		Timer:   metrics.NewRegisteredTimer("timer", nil),
	}
}

func (s *Status) Handler() servestk.Handler {
	if s.Counter == nil || s.Timer == nil {
		return func(w http.ResponseWriter, req *http.Request, f func(w http.ResponseWriter, req *http.Request)) {
			f(w, req)
		}
	}

	return func(w http.ResponseWriter, req *http.Request, f func(w http.ResponseWriter, req *http.Request)) {
		s.Counter.Inc(1)
		defer s.Counter.Dec(1)

		begin := time.Now()
		defer s.Timer.UpdateSince(begin)

		f(w, req)
	}
}

type StatusData struct {
	Processing   int64
	ResponseTime ResponseTime
	Throughput   Throughput
}

type ResponseTime struct {
	Max          int64
	Mean         float64
	Min          int64
	Percentile75 float64
	Percentile95 float64
	Percentile99 float64
	StdDev       float64
}

type Throughput struct {
	Count     int64
	Rate1min  float64
	Rate5min  float64
	Rate15min float64
	RateMean  float64
}

func (s *Status) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	d := &StatusData{
		Processing: s.Counter.Count(),
		ResponseTime: ResponseTime{
			Max:          s.Timer.Max(),
			Mean:         s.Timer.Mean(),
			Min:          s.Timer.Min(),
			Percentile75: s.Timer.Percentile(0.75),
			Percentile95: s.Timer.Percentile(0.95),
			Percentile99: s.Timer.Percentile(0.99),
			StdDev:       s.Timer.StdDev(),
		},
		Throughput: Throughput{
			Count:     s.Timer.Count(),
			Rate1min:  s.Timer.Rate1(),
			Rate5min:  s.Timer.Rate5(),
			Rate15min: s.Timer.Rate15(),
			RateMean:  s.Timer.RateMean(),
		},
	}

	b, err := json.MarshalIndent(d, "", "\t")
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.ReplyWith(w, 200, "application/json", b)
}
